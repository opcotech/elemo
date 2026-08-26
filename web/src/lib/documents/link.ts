import { collectedListQuery, cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1NamespacesDocumentsGetOptions } from "@/lib/api/query-options";
import { namespaceRefPath } from "@/lib/api/refs";
import { v1NamespacesDocumentsGet } from "@/lib/api/sdk";
import type { PartialDocument } from "@/lib/api/types";

export function filterAvailableDocuments<T extends { id: string }>(
  documents: readonly T[],
  alreadyRelatedIds: ReadonlySet<string>
): T[] {
  return documents.filter((document) => !alreadyRelatedIds.has(document.id));
}

export function relatedDocumentCatalogQueryOptions(
  organizationId: string,
  namespaceId: string
) {
  const listOptions = v1NamespacesDocumentsGetOptions({
    path: namespaceRefPath(organizationId, namespaceId),
    query: { all: true },
  });
  return collectedListQuery<PartialDocument>(
    listOptions,
    async (pageToken, signal) => {
      const { data } = await v1NamespacesDocumentsGet({
        path: namespaceRefPath(organizationId, namespaceId),
        query: {
          ...cursorPageQuery(pageToken),
          all: true,
        },
        signal,
        throwOnError: true,
      });
      return data;
    }
  );
}
