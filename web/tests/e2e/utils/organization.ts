import { getSession } from "./db";
import { generateXid } from "./xid";
import { getRandomString } from "./random";
import type { Organization, OrganizationStatus } from "@/lib/api";

export async function createDBOrganization(
  ownerId: string,
  status: OrganizationStatus = "active",
  overrides?: Partial<Organization>
): Promise<Organization> {
  const orgId = await generateXid();
  const createdAt = new Date().toISOString();

  const organization: Organization = {
    id: orgId,
    name: `${getRandomString(8)} Organization`,
    email: `${getRandomString(8)}-org@example.com`.toLowerCase(),
    logo: "https://picsum.photos/id/64/100/100",
    website: `https://${getRandomString(8).toLowerCase()}.example.com`,
    status: status,
    members: [ownerId],
    teams: [],
    namespaces: [],
    created_at: createdAt,
    updated_at: null,
    ...overrides,
  };

  const session = getSession();
  const membershipId = await generateXid();
  const permissionId = await generateXid();

  const resp = await session.executeWrite((tx: any) => {
    const query = `
      MATCH (u:User {id: $ownerId})
      CREATE (o:Organization {
        id: $id,
        name: $name,
        email: $email,
        logo: $logo,
        website: $website,
        status: $status,
        created_at: datetime($created_at)
      })
      CREATE (u)-[:MEMBER_OF {id: $membershipId, created_at: datetime($created_at)}]->(o)
      CREATE (u)-[:HAS_PERMISSION {id: $permissionId, created_at: datetime($created_at), kind: $permissionKind}]->(o)
      RETURN o
    `;

    return tx.run(query, {
      id: organization.id,
      name: organization.name,
      email: organization.email,
      logo: organization.logo,
      website: organization.website,
      status: organization.status,
      created_at: createdAt,
      ownerId: ownerId,
      membershipId: membershipId,
      permissionId: permissionId,
      permissionKind: "*", // All permissions
    });
  });

  const orgNode = resp.records[0].get("o").properties;
  await session.close();

  return {
    ...organization,
    id: orgNode.id,
    name: orgNode.name,
    email: orgNode.email,
    logo: orgNode.logo || null,
    website: orgNode.website || null,
    status: orgNode.status,
    created_at: orgNode.created_at?.toString() || createdAt,
  };
}

export async function createDBOrganizationWithPermission(
  ownerId: string,
  permissionKind: "*" | "create" | "write" | "read" | "delete",
  status: OrganizationStatus = "active",
  overrides?: Partial<Organization>
): Promise<Organization> {
  const orgId = await generateXid();
  const createdAt = new Date().toISOString();

  const organization: Organization = {
    id: orgId,
    name: `${getRandomString(8)} Organization`,
    email: `${getRandomString(8)}-org@example.com`.toLowerCase(),
    logo: "https://picsum.photos/id/64/100/100",
    website: `https://${getRandomString(8).toLowerCase()}.example.com`,
    status: status,
    members: [ownerId],
    teams: [],
    namespaces: [],
    created_at: createdAt,
    updated_at: null,
    ...overrides,
  };

  const session = getSession();
  const membershipId = await generateXid();
  const permissionId = await generateXid();

  const resp = await session.executeWrite((tx: any) => {
    const query = `
      MATCH (u:User {id: $ownerId})
      CREATE (o:Organization {
        id: $id,
        name: $name,
        email: $email,
        logo: $logo,
        website: $website,
        status: $status,
        created_at: datetime($created_at)
      })
      CREATE (u)-[:MEMBER_OF {id: $membershipId, created_at: datetime($created_at)}]->(o)
      CREATE (u)-[:HAS_PERMISSION {id: $permissionId, created_at: datetime($created_at), kind: $permissionKind}]->(o)
      RETURN o
    `;

    return tx.run(query, {
      id: organization.id,
      name: organization.name,
      email: organization.email,
      logo: organization.logo,
      website: organization.website,
      status: organization.status,
      created_at: createdAt,
      ownerId: ownerId,
      membershipId: membershipId,
      permissionId: permissionId,
      permissionKind: permissionKind,
    });
  });

  const orgNode = resp.records[0].get("o").properties;
  await session.close();

  return {
    ...organization,
    id: orgNode.id,
    name: orgNode.name,
    email: orgNode.email,
    logo: orgNode.logo || null,
    website: orgNode.website || null,
    status: orgNode.status,
    created_at: orgNode.created_at?.toString() || createdAt,
  };
}

/**
 * Adds a member to an organization with specific permissions.
 * This simulates adding a user as a member (not owner) to an existing organization.
 */
export async function addMemberToOrganization(
  orgId: string,
  memberId: string,
  permissionKind: "*" | "create" | "write" | "read" | "delete"
): Promise<void> {
  const session = getSession();
  const membershipId = await generateXid();
  const permissionId = await generateXid();
  const createdAt = new Date().toISOString();

  await session.executeWrite((tx: any) => {
    const query = `
      MATCH (u:User {id: $memberId})
      MATCH (o:Organization {id: $orgId})
      MERGE (u)-[:MEMBER_OF {id: $membershipId, created_at: datetime($created_at)}]->(o)
      WITH u, o
      MERGE (u)-[:HAS_PERMISSION {id: $permissionId, created_at: datetime($created_at), kind: $permissionKind}]->(o)
    `;

    return tx.run(query, {
      orgId,
      memberId,
      membershipId,
      permissionId,
      permissionKind,
      created_at: createdAt,
    });
  });

  await session.close();
}

/**
 * Creates a role for an organization and adds a creator as a member.
 */
export async function createDBRole(
  orgId: string,
  creatorId: string,
  roleName: string,
  description?: string
): Promise<string> {
  const session = getSession();
  const roleId = await generateXid();
  const membershipId = await generateXid();
  const hasTeamId = await generateXid();
  const permissionId = await generateXid();
  const createdAt = new Date().toISOString();

  await session.executeWrite((tx: any) => {
    const query = `
      MATCH (u:User {id: $creatorId})
      MATCH (o:Organization {id: $orgId})
      CREATE (r:Role {id: $roleId, name: $roleName, description: $description, created_at: datetime($created_at)})
      CREATE (o)-[:HAS_TEAM {id: $hasTeamId, created_at: datetime($created_at)}]->(r)
      CREATE (u)-[:MEMBER_OF {id: $membershipId, created_at: datetime($created_at)}]->(r)
      CREATE (u)-[:HAS_PERMISSION {id: $permissionId, created_at: datetime($created_at), kind: $permissionKind}]->(r)
    `;

    return tx.run(query, {
      orgId,
      creatorId,
      roleId,
      roleName,
      description: description || null,
      membershipId,
      hasTeamId,
      permissionId,
      permissionKind: "*",
      created_at: createdAt,
    });
  });

  await session.close();
  return roleId;
}

/**
 * Adds a member to a role.
 */
export async function addMemberToRole(
  roleId: string,
  memberId: string,
  orgId: string
): Promise<void> {
  const session = getSession();
  const membershipId = await generateXid();
  const createdAt = new Date().toISOString();

  await session.executeWrite((tx: any) => {
    const query = `
      MATCH (r:Role {id: $roleId})
      MATCH (u:User {id: $memberId})
      MATCH (o:Organization {id: $orgId})
      MERGE (u)-[:MEMBER_OF {id: $membershipId, created_at: datetime($created_at)}]->(r)
    `;

    return tx.run(query, {
      roleId,
      memberId,
      orgId,
      membershipId,
      created_at: createdAt,
    });
  });

  await session.close();
}
