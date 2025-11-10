import neo4j from "neo4j-driver";
import type { Driver } from "neo4j-driver";

import { USER_DEFAULT_PASSWORD, USER_DEFAULT_PASSWORD_HASH } from "./auth";
import { getRandomString } from "./random";
import type { getTestConfig } from "./test-config";
import { generateXid } from "./xid";

import type { PermissionKind, ResourceType, User } from "@/lib/api";

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
 * This function is used to bypass user invitation flow, mimicing a exsisting
 * condition in the system.
 */
export async function createUser(
  config: ReturnType<typeof getTestConfig>,
  overrides?: Partial<User & { password?: string }>
) {
  const driver = getDriver(config);
  const session = driver.session();
  let user: (User & { password?: string }) | null;

  const values = {
    id: await generateXid(),
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
 * Grant a permission to a user for a specific resource.
 *
 * The only acceptable use of this function is to mimic a specific permission
 * setup for a user. This function should only be used in scenarios where we
 * cannot use the API or we need to bypass the API for some reason.
 *
 * @param config - Test configuration
 * @param email - User email
 * @param resourceType - Resource type
 * @param resourceId - Resource ID
 * @param permissionKind - Permission kind
 */
export async function grantPermissionToUser(
  config: ReturnType<typeof getTestConfig>,
  email: string,
  resourceType: ResourceType,
  resourceId: string,
  permissionKind: PermissionKind
) {
  const driver = getDriver(config);
  const session = driver.session();

  const permissionId = await generateXid();
  const createdAt = new Date().toISOString();

  const query = `
    MATCH (u:User {email: $email})
    MATCH (t:${resourceType} {id: $resourceId})
    MERGE (u)-[p:HAS_PERMISSION {kind: $permissionKind}]->(t)
      ON CREATE SET p.id = $permissionId, p.created_at = datetime($createdAt)
  `;

  try {
    await session.executeWrite(async (tx) => {
      await tx.run(query, {
        email,
        resourceId,
        permissionKind,
        permissionId,
        createdAt,
      });
    });
  } catch (error) {
    console.error("Error granting permission to user", error);
  } finally {
    await session.close();
  }
}

/**
 * Grant a permission to a user for a specific resource.
 *
 * The only acceptable use of this function is to mimic a specific permission
 * setup for a user. This function should only be used in scenarios where we
 * cannot use the API or we need to bypass the API for some reason.
 *
 * @param config - Test configuration
 * @param email - User email
 * @param resourceType - Resource type
 * @param permissionKind - Permission kind
 */
export async function grantSystemPermissionToUser(
  config: ReturnType<typeof getTestConfig>,
  email: string,
  resourceType: ResourceType,
  permissionKind: PermissionKind
) {
  const driver = getDriver(config);
  const session = driver.session();

  const permissionId = await generateXid();
  const createdAt = new Date().toISOString();

  const query = `
    MATCH (u:User {email: $email})
    MATCH (rt:ResourceType {id: $resourceType})
    MERGE (u)-[p:HAS_PERMISSION {kind: $permissionKind}]->(rt)
      ON CREATE SET p.id = $permissionId, p.created_at = datetime($createdAt)
  `;

  try {
    await session.executeWrite(async (tx) => {
      await tx.run(query, {
        email,
        resourceType,
        permissionKind,
        permissionId,
        createdAt,
      });
    });
  } catch (error) {
    console.error("Error granting system permission to user", error);
  } finally {
    await session.close();
  }
}

/**
 * Grant system owner membership to a user.
 *
 * @param config - Test configuration
 * @param email - User email
 */
export async function grantSystemOwnerMembershipToUser(
  config: ReturnType<typeof getTestConfig>,
  email: string
) {
  const driver = getDriver(config);
  const session = driver.session();
  const query = `
    MATCH (u:User {email: $email})
    MATCH (r:Role {id: 'Owner'})
    MERGE (u)-[m:MEMBER_OF {id: $membershipId}]->(r)
      ON CREATE SET m.created_at = datetime()
  `;

  try {
    await session.executeWrite(async (tx) => {
      await tx.run(query, { email, membershipId: await generateXid() });
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
 *
 * @param config - Test configuration
 * @param email - User email
 * @param resourceType - Resource type
 * @param resourceId - Resource ID
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
        membershipId: await generateXid(),
      });
    });
  } finally {
    await session.close();
  }
}

/**
 * Ensure system owner user exists in the database.
 * Creates the user if it doesn't exist (idempotent).
 */
export async function ensureSystemOwner(
  config: ReturnType<typeof getTestConfig>
) {
  const driver = getDriver(config);

  try {
    const session = driver.session();

    try {
      // Check if user already exists
      const checkResult = await session.executeRead(async (tx) => {
        return await tx.run("MATCH (u:User {id: $userId}) RETURN u", {
          userId: "d49pd9v92rs4hfc796k0",
        });
      });

      if (checkResult.records.length > 0) {
        console.log("System owner user already exists, skipping creation");
        return;
      }

      await createUser(config, {
        username: "e2e-test-owner",
        first_name: "E2E Test",
        last_name: "Owner",
        email: config.systemOwnerEmail,
        password: USER_DEFAULT_PASSWORD,
      });

      await grantSystemOwnerMembershipToUser(config, config.systemOwnerEmail);

      console.debug("System owner user created successfully");
    } finally {
      await session.close();
    }
  } finally {
    await driver.close();
  }
}
