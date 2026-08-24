import {
  createGrant,
  createOrganization,
  createRole,
  deleteGrant,
  getEffectiveActions,
} from "./api";
import { expect, test } from "./fixtures";
import { USER_DEFAULT_PASSWORD } from "./utils/auth";
import {
  createUser,
  grantMembershipToUser,
  grantOrganizationCreateToUser,
} from "./utils/db";
import { getRandomString } from "./utils/random";

import { v1OrganizationsCreate, v1PermissionResourceGet } from "@/lib/api/sdk";

test.describe("@permissions.grants Scoped ReBAC Grant E2E Tests", () => {
  test("should assign a role to a user principal at organization scope", async ({
    testConfig,
    createApiClient,
  }) => {
    const owner = await createUser(testConfig);
    const member = await createUser(testConfig);
    await grantOrganizationCreateToUser(testConfig, owner.email);

    const ownerClient = await createApiClient(
      owner.email,
      USER_DEFAULT_PASSWORD
    );
    const uniqueId = getRandomString(8);
    const organization = await createOrganization(ownerClient, {
      name: `Grant Org ${uniqueId}`,
      email: `grant-${uniqueId}@example.com`,
    });

    await grantMembershipToUser(
      testConfig,
      member.email,
      "Organization",
      organization.id
    );

    const role = await createRole(ownerClient, organization.id, {
      name: `Org Editors ${uniqueId}`,
      actions: ["organization.update"],
    });

    const grant = await createGrant(ownerClient, {
      principal: { resourceType: "User", id: member.id },
      scope: { resourceType: "Organization", id: organization.id },
      role_id: role.id,
    });
    expect(grant.id).toBeTruthy();

    const memberClient = await createApiClient(
      member.email,
      USER_DEFAULT_PASSWORD
    );
    const effective = await getEffectiveActions(
      memberClient,
      "Organization",
      organization.id
    );
    expect(effective.actions).toContain("organization.read");
    expect(effective.actions).toContain("organization.update");

    await deleteGrant(ownerClient, grant.id);

    const afterRevoke = await getEffectiveActions(
      memberClient,
      "Organization",
      organization.id
    );
    expect(afterRevoke.actions).toContain("organization.read");
    expect(afterRevoke.actions).not.toContain("organization.update");
  });

  test("should assign a role to an organization principal at a scope", async ({
    testConfig,
    createApiClient,
  }) => {
    const owner = await createUser(testConfig);
    const member = await createUser(testConfig);
    await grantOrganizationCreateToUser(testConfig, owner.email);

    const ownerClient = await createApiClient(
      owner.email,
      USER_DEFAULT_PASSWORD
    );
    const uniqueId = getRandomString(8);
    const organization = await createOrganization(ownerClient, {
      name: `Org Principal ${uniqueId}`,
      email: `org-principal-${uniqueId}@example.com`,
    });

    await grantMembershipToUser(
      testConfig,
      member.email,
      "Organization",
      organization.id
    );

    const role = await createRole(ownerClient, organization.id, {
      name: `Org Updaters ${uniqueId}`,
      actions: ["organization.update"],
    });

    await createGrant(ownerClient, {
      principal: { resourceType: "Organization", id: organization.id },
      scope: { resourceType: "Organization", id: organization.id },
      role_id: role.id,
    });

    const memberClient = await createApiClient(
      member.email,
      USER_DEFAULT_PASSWORD
    );
    const effective = await getEffectiveActions(
      memberClient,
      "Organization",
      organization.id
    );
    expect(effective.actions).toContain("organization.update");
  });

  test("should deny organization.create for a user without an Installation grant", async ({
    testConfig,
    createApiClient,
  }) => {
    const user = await createUser(testConfig);
    const client = await createApiClient(user.email, USER_DEFAULT_PASSWORD);
    const uniqueId = getRandomString(8);

    const result = await v1OrganizationsCreate({
      client,
      body: {
        name: `Denied Org ${uniqueId}`,
        email: `denied-${uniqueId}@example.com`,
      },
    });

    expect(result.error).toBeTruthy();
    expect(result.response?.status).toBe(403);
  });

  test("should return effective actions as a string array", async ({
    testConfig,
    createApiClient,
  }) => {
    const owner = await createUser(testConfig);
    await grantOrganizationCreateToUser(testConfig, owner.email);

    const ownerClient = await createApiClient(
      owner.email,
      USER_DEFAULT_PASSWORD
    );
    const uniqueId = getRandomString(8);
    const organization = await createOrganization(ownerClient, {
      name: `Actions Org ${uniqueId}`,
      email: `actions-${uniqueId}@example.com`,
    });

    const result = await v1PermissionResourceGet({
      client: ownerClient,
      path: { resourceId: `Organization:${organization.id}` },
      throwOnError: true,
    });

    expect(Array.isArray(result.data.actions)).toBeTruthy();
    expect(result.data.actions).toContain("organization.read");
    expect(result.data.actions).toContain("role.manage");
    expect(result.data.actions).not.toContain("*");
  });
});
