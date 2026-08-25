import { withErrorHandling } from "./error-handler";

import type { Client } from "@/lib/api/client";
import {
  v1NamespacesFoldersCreate,
  v1OrganizationsFoldersCreate,
} from "@/lib/api/sdk";
import type { Folder, FolderCreate } from "@/lib/api/types";

export async function createNamespaceFolder(
  client: Client,
  organizationId: string,
  namespaceId: string,
  body: FolderCreate
): Promise<Folder> {
  const response = await withErrorHandling(
    async () => {
      return await v1NamespacesFoldersCreate({
        client,
        path: { organizationRef: organizationId, namespaceRef: namespaceId },
        body,
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${organizationId}/namespaces/${namespaceId}/folders`,
      method: "POST",
    }
  );

  return response.data;
}

export async function createOrganizationFolder(
  client: Client,
  organizationId: string,
  body: FolderCreate
): Promise<Folder> {
  const response = await withErrorHandling(
    async () => {
      return await v1OrganizationsFoldersCreate({
        client,
        path: { organizationRef: organizationId },
        body,
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${organizationId}/folders`,
      method: "POST",
    }
  );

  return response.data;
}
