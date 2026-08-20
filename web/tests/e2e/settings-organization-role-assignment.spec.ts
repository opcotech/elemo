import {
  createGrant,
  createOrganization,
  createRole,
  getEffectiveActions,
} from "./api";
import { expect, test } from "./fixtures";
import { USER_DEFAULT_PASSWORD } from "./utils/auth";
import {
  createUser,
  grantActionsToUser,
  grantMembershipToUser,
  grantOrganizationCreateToUser,
} from "./utils/db";
import { getRandomString } from "./utils/random";

import type { User } from "@/lib/api/types";

test.describe("@settings.organization-role-assignment Role Grant Assignment E2E Tests", () => {
  let owner: User;
  let member: User;
  let organizationId: string;

  test.beforeAll(async ({ testConfig, createApiClient }) => {
    owner = await createUser(testConfig);
    member = await createUser(testConfig);

    await grantOrganizationCreateToUser(testConfig, owner.email);

    const uniqueId = getRandomString(8);
    const apiClient = await createApiClient(owner.email, USER_DEFAULT_PASSWORD);
    const organization = await createOrganization(apiClient, {
      name: `Test Org Role Assignment ${uniqueId}`,
      email: `test-role-assignment-${uniqueId}@example.com`,
    });
    organizationId = organization.id;

    await grantMembershipToUser(
      testConfig,
      member.email,
      "Organization",
      organizationId
    );
    await grantActionsToUser(
      testConfig,
      member.email,
      "Organization",
      organizationId,
      ["organization.read"]
    );
  });

  test("should grant a role bundle to a user at organization scope", async ({
    createApiClient,
  }) => {
    const apiClient = await createApiClient(owner.email, USER_DEFAULT_PASSWORD);
    const role = await createRole(apiClient, organizationId, {
      name: `Assignment Role ${getRandomString(8)}`,
      actions: ["organization.update"],
    });

    await createGrant(apiClient, {
      principal: { resourceType: "User", id: member.id },
      scope: { resourceType: "Organization", id: organizationId },
      role_id: role.id,
    });

    const memberClient = await createApiClient(
      member.email,
      USER_DEFAULT_PASSWORD
    );
    const effective = await getEffectiveActions(
      memberClient,
      "Organization",
      organizationId
    );
    expect(effective.actions).toContain("organization.update");
  });

  test("should grant a role bundle to the organization principal", async ({
    createApiClient,
  }) => {
    const apiClient = await createApiClient(owner.email, USER_DEFAULT_PASSWORD);
    const role = await createRole(apiClient, organizationId, {
      name: `Org Principal Role ${getRandomString(8)}`,
      actions: ["namespace.read"],
    });

    await createGrant(apiClient, {
      principal: { resourceType: "Organization", id: organizationId },
      scope: { resourceType: "Organization", id: organizationId },
      role_id: role.id,
    });

    const memberClient = await createApiClient(
      member.email,
      USER_DEFAULT_PASSWORD
    );
    const effective = await getEffectiveActions(
      memberClient,
      "Organization",
      organizationId
    );
    expect(effective.actions).toContain("namespace.read");
  });
});
