import type { Query, QueryClient } from "@tanstack/react-query";

import { v1DocumentGetOptions } from "@/lib/api/query-options";

/** Hey-api `_id` values for parent document list queries. */
export const DOCUMENT_LIST_QUERY_IDS = [
  "v1NamespacesDocumentsGet",
  "v1ProjectsDocumentsGet",
  "v1OrganizationsDocumentsGet",
  "v1IssuesDocumentsGet",
] as const;

const FOLDER_LIST_QUERY_IDS = [
  "v1NamespacesFoldersGet",
  "v1OrganizationsFoldersGet",
] as const;

const documentListQueryIds = new Set<string>(DOCUMENT_LIST_QUERY_IDS);
const folderListQueryIds = new Set<string>(FOLDER_LIST_QUERY_IDS);

function queryOperationId(query: Query): string | undefined {
  const key = query.queryKey[0];
  if (!key || typeof key !== "object") {
    return undefined;
  }
  const id = (key as { _id?: string })._id;
  return typeof id === "string" ? id : undefined;
}

export function isDocumentListQuery(query: Query): boolean {
  const id = queryOperationId(query);
  return typeof id === "string" && documentListQueryIds.has(id);
}

export function isFolderListQuery(query: Query): boolean {
  const id = queryOperationId(query);
  return typeof id === "string" && folderListQueryIds.has(id);
}

export async function invalidateDocumentQueries(
  queryClient: QueryClient,
  documentId: string
): Promise<void> {
  await Promise.all([
    queryClient.invalidateQueries({
      queryKey: v1DocumentGetOptions({
        path: { id: documentId },
      }).queryKey,
    }),
    queryClient.invalidateQueries({
      predicate: isDocumentListQuery,
    }),
    queryClient.invalidateQueries({
      predicate: isFolderListQuery,
    }),
    queryClient.invalidateQueries({ queryKey: ["library-folder-options"] }),
    queryClient.invalidateQueries({ queryKey: ["folder-path"] }),
  ]);
}

export async function invalidateLibraryQueries(
  queryClient: QueryClient
): Promise<void> {
  await Promise.all([
    queryClient.invalidateQueries({
      predicate: isDocumentListQuery,
    }),
    queryClient.invalidateQueries({
      predicate: isFolderListQuery,
    }),
    queryClient.invalidateQueries({ queryKey: ["library-folder-options"] }),
    queryClient.invalidateQueries({ queryKey: ["folder-path"] }),
  ]);
}
