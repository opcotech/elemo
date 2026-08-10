import { withErrorHandling } from "./error-handler";
import { getRandomString } from "../utils/random";

import type { Client } from "@/lib/client/client";
import { v1NamespacesProjectsCreate, v1ProjectGet } from "@/lib/client/sdk.gen";
import type { Project, ProjectCreate } from "@/lib/client/types.gen";

/**
 * Generate a project key that satisfies API constraints (3–6 uppercase ASCII letters).
 */
export function getRandomProjectKey(length: number = 6): string {
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
  const size = Math.min(Math.max(length, 3), 6);
  let result = "";
  for (let i = 0; i < size; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

/**
 * Create a project via API.
 *
 * @param client - Authenticated API client
 * @param namespaceId - Namespace ID to create the project in
 * @param projectData - Project data (key and name are required)
 * @returns Created project with ID
 */
export async function createProject(
  client: Client,
  namespaceId: string,
  projectData: Partial<ProjectCreate> & { key: string; name: string }
): Promise<Project> {
  const projectCreateData: ProjectCreate = {
    key: projectData.key,
    name: projectData.name,
    description:
      projectData.description ?? `Project description ${getRandomString(8)}`,
    logo: projectData.logo,
    status: projectData.status,
  };

  const response = await withErrorHandling(
    async () => {
      return await v1NamespacesProjectsCreate({
        client,
        path: { id: namespaceId },
        body: projectCreateData,
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/namespaces/${namespaceId}/projects`,
      method: "POST",
    }
  );

  const projectId = response.data.id || "";

  const projectResponse = await withErrorHandling(
    async () => {
      return await v1ProjectGet({
        client,
        path: { id: projectId },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/projects/${projectId}`,
      method: "GET",
    }
  );

  return projectResponse.data;
}
