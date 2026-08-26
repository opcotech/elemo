import {
  createAuthenticatedClient,
  createOrganization,
  createProject,
  getRandomProjectKey,
  withErrorHandling,
} from "../api";
import { USER_DEFAULT_PASSWORD } from "../utils/auth";
import { createUser, grantOrganizationCreateToUser } from "../utils/db";
import { getRandomSlug, getRandomString } from "../utils/random";
import type { TestConfig } from "../utils/test-config";

import type { Client } from "@/lib/api/client";
import { v1OrganizationsNamespacesCreate } from "@/lib/api/sdk";
import type { User } from "@/lib/api/types";

export interface OwnerWorkspace {
  owner: User;
  client: Client;
  organizationId: string;
  organizationSlug: string;
  organizationName: string;
  namespaceId: string;
  namespaceSlug: string;
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
  options?: {
    namePrefix?: string;
    organizationSlug?: string;
    namespaceSlug?: string;
    projectKey?: string;
  }
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
  const organizationSlug = options?.organizationSlug ?? getRandomSlug("org");
  const organization = await createOrganization(client, {
    slug: organizationSlug,
    name: organizationName,
    email: `${prefix.toLowerCase().replaceAll(/\s+/g, "-")}-${suffix}@example.com`,
  });

  const namespaceName = `${prefix} Namespace ${suffix}`;
  const namespaceSlug = options?.namespaceSlug ?? getRandomSlug("ns");
  const namespaceResponse = await withErrorHandling(
    async () => {
      return await v1OrganizationsNamespacesCreate({
        client,
        path: { organizationRef: organization.id },
        body: {
          name: namespaceName,
          slug: namespaceSlug,
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

  const projectKey = options?.projectKey ?? getRandomProjectKey();
  const projectName = `${prefix} Project ${suffix}`;
  const project = await createProject(
    client,
    organization.id,
    namespaceResponse.data.id ?? "",
    {
      key: projectKey,
      name: projectName,
      description: `Project for ${prefix} ${suffix}`,
    }
  );

  return {
    owner,
    client,
    organizationId: organization.id,
    organizationSlug: organization.slug,
    organizationName,
    namespaceId: namespaceResponse.data.id ?? "",
    namespaceSlug,
    namespaceName,
    projectId: project.id,
    projectName,
    projectKey,
  };
}
