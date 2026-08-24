import {
  createGrant,
  createOrganization,
  createProject,
  getRandomProjectKey,
} from "./api";
import { expect, test } from "./fixtures";
import { USER_DEFAULT_PASSWORD } from "./utils/auth";
import { createUser, grantOrganizationCreateToUser } from "./utils/db";
import { getRandomString } from "./utils/random";

import {
  v1NamespaceGet,
  v1NamespacesGet,
  v1NamespacesProjectsGet,
  v1OrganizationGet,
  v1OrganizationsNamespacesCreate,
  v1OrganizationsNamespacesGet,
  v1ProjectGet,
} from "@/lib/api/sdk";

const PROJECT_VIEWER_ACTIONS = ["project.read", "issue.read", "document.read"];

test.describe("@permissions.cross-org Cross-organization sharing E2E Tests", () => {
  test("org B principal can read a shared project in org A but not org A or sibling projects", async ({
    testConfig,
    createApiClient,
  }) => {
    const userA = await createUser(testConfig);
    const userB = await createUser(testConfig);
    await grantOrganizationCreateToUser(testConfig, userA.email);
    await grantOrganizationCreateToUser(testConfig, userB.email);

    const clientA = await createApiClient(userA.email, USER_DEFAULT_PASSWORD);
    const clientB = await createApiClient(userB.email, USER_DEFAULT_PASSWORD);

    const uniqueId = getRandomString(8);
    const orgA = await createOrganization(clientA, {
      name: `Org A ${uniqueId}`,
      email: `org-a-${uniqueId}@example.com`,
    });
    const orgB = await createOrganization(clientB, {
      name: `Org B ${uniqueId}`,
      email: `org-b-${uniqueId}@example.com`,
    });

    const namespace = await v1OrganizationsNamespacesCreate({
      client: clientA,
      path: { id: orgA.id },
      body: {
        name: `NS ${uniqueId}`,
        description: `Namespace ${uniqueId}`,
      },
      throwOnError: true,
    });
    const namespaceId = namespace.data.id ?? "";

    const sharedProject = await createProject(clientA, namespaceId, {
      key: getRandomProjectKey(),
      name: `Shared ${uniqueId}`,
    });
    const siblingProject = await createProject(clientA, namespaceId, {
      key: getRandomProjectKey(),
      name: `Sibling ${uniqueId}`,
    });

    await createGrant(clientA, {
      principal: { resourceType: "Organization", id: orgB.id },
      scope: { resourceType: "Project", id: sharedProject.id },
      actions: PROJECT_VIEWER_ACTIONS,
    });

    const shared = await v1ProjectGet({
      client: clientB,
      path: { id: sharedProject.id },
      throwOnError: true,
    });
    expect(shared.data.id).toBe(sharedProject.id);

    const orgAResult = await v1OrganizationGet({
      client: clientB,
      path: { id: orgA.id },
    });
    expect(orgAResult.error).toBeTruthy();
    expect(orgAResult.response?.status).toBe(403);

    const siblingResult = await v1ProjectGet({
      client: clientB,
      path: { id: siblingProject.id },
    });
    expect(siblingResult.error).toBeTruthy();
    expect(siblingResult.response?.status).toBe(403);

    const reachable = await v1NamespacesGet({
      client: clientB,
      throwOnError: true,
    });
    expect(reachable.data.items.map((item) => item.id)).toContain(namespaceId);
    const reachableNamespace = reachable.data.items.find(
      (item) => item.id === namespaceId
    );
    expect(reachableNamespace?.organization.id).toBe(orgA.id);
    expect(reachableNamespace?.organization.name).toBe(orgA.name);

    const listedNamespaces = await v1OrganizationsNamespacesGet({
      client: clientB,
      path: { id: orgA.id },
      throwOnError: true,
    });
    expect(listedNamespaces.data.items.map((item) => item.id)).toEqual([
      namespaceId,
    ]);

    const namespaceGet = await v1NamespaceGet({
      client: clientB,
      path: { id: namespaceId },
    });
    expect(namespaceGet.error).toBeTruthy();
    expect(namespaceGet.response?.status).toBe(403);

    const projects = await v1NamespacesProjectsGet({
      client: clientB,
      path: { id: namespaceId },
      throwOnError: true,
    });
    expect(projects.data.items.map((item) => item.id)).toEqual([
      sharedProject.id,
    ]);
  });
});
