import { withErrorHandling } from "./error-handler";

import type { Client } from "@/lib/api/client";
import {
  v1NamespacesFoldersCreate,
  v1OrganizationsFoldersCreate,
} from "@/lib/api/sdk";
import type { Folder, FolderCreate } from "@/lib/api/types";

async function createFolder(
  client: Client,
  create:
    typeof v1NamespacesFoldersCreate | typeof v1OrganizationsFoldersCreate,
  ownerId: string,
  endpoint: string,
  body: FolderCreate
): Promise<Folder> {
  const response = await withErrorHandling(
    async () => {
      return await create({
        client,
        path: { id: ownerId },
        body,
        throwOnError: true,
      });
    },
    {
      endpoint,
      method: "POST",
    }
  );

  return response.data;
}

export async function createNamespaceFolder(
  client: Client,
  namespaceId: string,
  body: FolderCreate
): Promise<Folder> {
  return createFolder(
    client,
    v1NamespacesFoldersCreate,
    namespaceId,
    `/v1/namespaces/${namespaceId}/folders`,
    body
  );
}

export async function createOrganizationFolder(
  client: Client,
  organizationId: string,
  body: FolderCreate
): Promise<Folder> {
  return createFolder(
    client,
    v1OrganizationsFoldersCreate,
    organizationId,
    `/v1/organizations/${organizationId}/folders`,
    body
  );
}
