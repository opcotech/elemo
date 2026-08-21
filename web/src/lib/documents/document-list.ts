import type { PartialDocument } from "@/lib/api/types";
import { getInitials } from "@/lib/utils";

export const ALL_DOCUMENT_CREATORS = "all";

export const documentListSorts = [
  "updated-desc",
  "created-desc",
  "created-asc",
  "title-asc",
] as const;

export type DocumentListSort = (typeof documentListSorts)[number];

export const documentListSortLabels: Record<DocumentListSort, string> = {
  "updated-desc": "Updated",
  "created-desc": "Created",
  "created-asc": "Oldest",
  "title-asc": "Title",
};

export function isDocumentListSort(value: string): value is DocumentListSort {
  return (documentListSorts as readonly string[]).includes(value);
}

export function documentAuthorName(
  document: Pick<PartialDocument, "created_by">
): string {
  return (
    `${document.created_by.first_name} ${document.created_by.last_name}`.trim() ||
    document.created_by.id
  );
}

export function documentExcerpt(
  document: Pick<PartialDocument, "excerpt">
): string | null {
  const excerpt = document.excerpt?.trim();
  return excerpt ? excerpt : null;
}

export function documentUpdatedAt(
  document: Pick<PartialDocument, "created_at" | "updated_at">
): string | null {
  return document.updated_at ?? document.created_at ?? null;
}

export interface DocumentCreatorOption {
  id: string;
  name: string;
  picture: string | null;
  initials: string;
}

export function documentCreators(
  documents: readonly PartialDocument[]
): DocumentCreatorOption[] {
  const byId = new Map<string, DocumentCreatorOption>();
  for (const document of documents) {
    const user = document.created_by;
    if (!byId.has(user.id)) {
      byId.set(user.id, {
        id: user.id,
        name: documentAuthorName(document),
        picture: user.picture ?? null,
        initials: getInitials(user.first_name, user.last_name),
      });
    }
  }
  return [...byId.values()].sort((left, right) =>
    left.name.localeCompare(right.name)
  );
}

export function filterDocuments(
  documents: readonly PartialDocument[],
  query: string,
  creatorId?: string | null
): PartialDocument[] {
  const normalized = query.trim().toLowerCase();
  const creatorFilter =
    creatorId && creatorId !== ALL_DOCUMENT_CREATORS ? creatorId : null;

  return documents.filter((document) => {
    if (creatorFilter && document.created_by.id !== creatorFilter) {
      return false;
    }
    if (!normalized) {
      return true;
    }
    return [
      document.title,
      documentExcerpt(document),
      documentAuthorName(document),
    ]
      .filter(Boolean)
      .join(" ")
      .toLowerCase()
      .includes(normalized);
  });
}

export function sortDocuments(
  documents: readonly PartialDocument[],
  sort: DocumentListSort
): PartialDocument[] {
  const items = [...documents];
  items.sort((left, right) => {
    if (sort === "title-asc") {
      return left.title.localeCompare(right.title);
    }

    if (sort === "updated-desc") {
      const updatedDelta =
        documentTime(documentUpdatedAt(right)) -
        documentTime(documentUpdatedAt(left));
      return updatedDelta !== 0
        ? updatedDelta
        : left.title.localeCompare(right.title);
    }

    const leftTime = documentTime(left.created_at);
    const rightTime = documentTime(right.created_at);
    const createdDelta =
      sort === "created-asc" ? leftTime - rightTime : rightTime - leftTime;
    return createdDelta !== 0
      ? createdDelta
      : left.title.localeCompare(right.title);
  });
  return items;
}

export function visibleDocuments(
  documents: readonly PartialDocument[],
  query: string,
  sort: DocumentListSort,
  creatorId?: string | null
): PartialDocument[] {
  return sortDocuments(filterDocuments(documents, query, creatorId), sort);
}

function documentTime(value: string | null | undefined): number {
  if (!value) {
    return 0;
  }
  const time = new Date(value).getTime();
  return Number.isNaN(time) ? 0 : time;
}
