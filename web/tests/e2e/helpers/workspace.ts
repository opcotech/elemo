import {
  createAuthenticatedClient,
  createOrganization,
  createProject,
  getRandomProjectKey,
  withErrorHandling,
} from "../api";
import { USER_DEFAULT_PASSWORD } from "../utils/auth";
import { createUser, grantOrganizationCreateToUser } from "../utils/db";
import { getRandomString } from "../utils/random";
import type { TestConfig } from "../utils/test-config";

import type { Client } from "@/lib/api/client";
import { v1OrganizationsNamespacesCreate } from "@/lib/api/sdk";
import type { User } from "@/lib/api/types";

export interface OwnerWorkspace {
  owner: User;
  client: Client;
  organizationId: string;
  organizationName: string;
  namespaceId: string;
  namespaceName: string;
  projectId: string;
  projectName: string;
  projectKey: string;
}

/**
 * Seed an owner user plus org, namespace, and project for issue E2E specs.
 */
export async function seedOwnerWorkspace(
  testConfig: TestConfig,
  options?: { namePrefix?: string }
): Promise<OwnerWorkspace> {
  const prefix = options?.namePrefix ?? "Workspace";
  const suffix = getRandomString(8);

  const owner = await createUser(testConfig);
  await grantOrganizationCreateToUser(testConfig, owner.email);

  const client = await createAuthenticatedClient(
    owner.email,
    USER_DEFAULT_PASSWORD
  );

  const organizationName = `${prefix} Org ${suffix}`;
  const organization = await createOrganization(client, {
    name: organizationName,
    email: `${prefix.toLowerCase().replaceAll(/\s+/g, "-")}-${suffix}@example.com`,
  });

  const namespaceName = `${prefix} Namespace ${suffix}`;
  const namespaceResponse = await withErrorHandling(
    async () => {
      return await v1OrganizationsNamespacesCreate({
        client,
        path: { id: organization.id },
        body: {
          name: namespaceName,
          description: `Namespace for ${prefix} ${suffix}`,
        },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${organization.id}/namespaces`,
      method: "POST",
    }
  );

  const projectKey = getRandomProjectKey();
  const projectName = `${prefix} Project ${suffix}`;
  const project = await createProject(client, namespaceResponse.data.id ?? "", {
    key: projectKey,
    name: projectName,
    description: `Project for ${prefix} ${suffix}`,
  });

  return {
    owner,
    client,
    organizationId: organization.id,
    organizationName,
    namespaceId: namespaceResponse.data.id ?? "",
    namespaceName,
    projectId: project.id,
    projectName,
    projectKey,
  };
}
