import neo4j from "neo4j-driver";
import type { Driver } from "neo4j-driver";

import { USER_DEFAULT_PASSWORD_HASH } from "./auth";
import { getRandomString } from "./random";
import type { getTestConfig } from "./test-config";
import { generateXid } from "./xid";

import type { ResourceType, User } from "@/lib/api/types";

/** Stable Installation node id used as the organization.create scope. */
export const INSTALLATION_ID = "00000000000000000000";

/** Stable e2e privileged user id (organization.create on Installation). */
export const PRIVILEGED_USER_ID = "d49pd9v92rs4hfc796k0";

let _cachedDriver: Driver | null = null;

function getDriver(config: ReturnType<typeof getTestConfig>) {
  if (_cachedDriver) {
    return _cachedDriver;
  }

  _cachedDriver = neo4j.driver(
    config.neo4jUrl,
    neo4j.auth.basic(config.neo4jUser, config.neo4jPassword)
  );

  return _cachedDriver;
}

/**
 * Create a new user in the database.
 *
 * This function is used to bypass user invitation flow, mimicking an existing
 * condition in the system. Users are labeled Principal so they can hold grants.
 */
export async function createUser(
  config: ReturnType<typeof getTestConfig>,
  overrides?: Partial<User & { password?: string }>
) {
  const driver = getDriver(config);
  const session = driver.session();
  let user: (User & { password?: string }) | null;

  const values = {
    id: generateXid(),
    email: `${getRandomString()}-test@example.com`.toLowerCase(),
    username: getRandomString(8).toLowerCase(),
    password: USER_DEFAULT_PASSWORD_HASH,
    status: "active",
    first_name: getRandomString(),
    last_name: getRandomString(),
    picture: "https://picsum.photos/id/1084/200/200.jpg?grayscale",
    title: "Poor Test User",
    bio: "I am a poor test user.",
    phone: "+12345678900",
    address: "2900 S Congress Ave, Austin, TX",
    links: ["https://example.com"],
    languages: ["en"],
    ...overrides,
  };

  const query = `
    MERGE (u:User {email: $email})
    ON CREATE SET u += {
      id: $id,
      username: $username,
      email: $email,
      password: $password,
      status: $status,
      first_name: $first_name,
      last_name: $last_name,
      picture: $picture,
      title: $title,
      bio: $bio,
      phone: $phone,
      address: $address,
      links: $links,
      languages: $languages,
      created_at: datetime()
    }
    SET u:Principal
    RETURN u
  `;

  try {
    user = await session.executeWrite(async (tx) => {
      const result = await tx.run(query, values);
      return result.records[0].get("u").properties;
    });
  } finally {
    await session.close();
  }

  return user as User;
}

/**
 * Grant actions to a user on a specific resource via a GRANTED edge.
 *
 * The only acceptable use of this function is to mimic a specific grant
 * setup for a user. Prefer the permissions API when the caller already has
 * permission.manage on the scope.
 */
export async function grantActionsToUser(
  config: ReturnType<typeof getTestConfig>,
  email: string,
  resourceType: ResourceType,
  resourceId: string,
  actions: string[]
) {
  const driver = getDriver(config);
  const session = driver.session();

  const grantId = generateXid();

  const query = `
    MATCH (u:User {email: $email})
    SET u:Principal
    WITH u
    MATCH (t:${resourceType} {id: $resourceId})
    CREATE (u)-[g:GRANTED {
      id: $grantId,
      actions: $actions,
      role_id: "",
      created_at: datetime()
    }]->(t)
    RETURN g
  `;

  try {
    const granted = await session.executeWrite(async (tx) => {
      const result = await tx.run(query, {
        email,
        resourceId,
        actions,
        grantId,
      });
      return result.records.length > 0;
    });
    if (!granted) {
      throw new Error(
        `Failed to grant [${actions.join(", ")}] on ${resourceType}:${resourceId} to ${email}`
      );
    }
  } finally {
    await session.close();
  }
}

/**
 * Grant organization.create on the Installation node to a user principal.
 */
export async function grantOrganizationCreateToUser(
  config: ReturnType<typeof getTestConfig>,
  email: string
) {
  const driver = getDriver(config);
  const session = driver.session();
  const query = `
    MATCH (u:User {email: $email})
    SET u:Principal
    WITH u
    MERGE (i:Installation {id: $installationId})
    ON CREATE SET i.created_at = datetime()
    MERGE (u)-[g:GRANTED]->(i)
    ON CREATE SET
      g.id = $grantId,
      g.actions = $actions,
      g.role_id = "",
      g.created_at = datetime()
    ON MATCH SET
      g.actions = $actions,
      g.updated_at = datetime()
  `;

  try {
    await session.executeWrite(async (tx) => {
      await tx.run(query, {
        email,
        installationId: INSTALLATION_ID,
        grantId: generateXid(),
        actions: ["organization.create", "plugin.install"],
      });
    });
  } finally {
    await session.close();
  }
}

/**
 * Grant membership to a user for a specific resource.
 *
 * The only acceptable use of this function is to mimic a specific membership
 * setup for a user. This function should only be used in scenarios where we
 * cannot use the API or we need to bypass the API for some reason.
 */
export async function grantMembershipToUser(
  config: ReturnType<typeof getTestConfig>,
  email: string,
  resourceType: ResourceType,
  resourceId: string
) {
  const driver = getDriver(config);
  const session = driver.session();
  const query = `
    MATCH (u:User {email: $email})
    SET u:Principal
    WITH u
    MATCH (t:${resourceType} {id: $resourceId})
    MERGE (u)-[m:MEMBER_OF {id: $membershipId}]->(t)
      ON CREATE SET m.created_at = datetime()
  `;

  try {
    await session.executeWrite(async (tx) => {
      await tx.run(query, {
        email,
        resourceType,
        resourceId,
        membershipId: generateXid(),
      });
    });
  } finally {
    await session.close();
  }
}

/**
 * Ensure the e2e privileged user exists with organization.create on Installation.
 * Creates the user if it doesn't exist (idempotent). Does not use system roles.
 */
export async function ensurePrivilegedUser(
  config: ReturnType<typeof getTestConfig>
) {
  const driver = getDriver(config);
  const session = driver.session();

  try {
    const checkResult = await session.executeRead(async (tx) => {
      return await tx.run("MATCH (u:User {id: $userId}) RETURN u", {
        userId: PRIVILEGED_USER_ID,
      });
    });

    if (checkResult.records.length === 0) {
      await createUser(config, {
        id: PRIVILEGED_USER_ID,
        username: "e2e-test-owner",
        first_name: "E2E Test",
        last_name: "Owner",
        email: config.systemOwnerEmail,
        password: USER_DEFAULT_PASSWORD_HASH,
      });
    }

    await grantOrganizationCreateToUser(config, config.systemOwnerEmail);

    console.debug("E2E privileged user is ready");
  } finally {
    await session.close();
  }
}
