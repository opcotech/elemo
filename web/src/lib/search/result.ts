import type { SearchResult } from "@/lib/api/types";
import type { AppEntityType } from "@/lib/entity-types";

export const SEARCH_RESOURCE_TYPES = [
  "Organization",
  "Namespace",
  "Project",
  "Issue",
  "Document",
] as const;

export type SearchResourceType = (typeof SEARCH_RESOURCE_TYPES)[number];

const searchEntityTypes: Record<SearchResourceType, AppEntityType> = {
  Organization: "organization",
  Namespace: "namespace",
  Project: "project",
  Issue: "work-item",
  Document: "document",
};

export const SEARCH_TYPE_LABELS: Record<"all" | SearchResourceType, string> = {
  all: "All types",
  Organization: "Organizations",
  Namespace: "Namespaces",
  Project: "Projects",
  Issue: "Issues",
  Document: "Documents",
};

export function searchResultEntityType(result: SearchResult): AppEntityType {
  return searchEntityTypes[result.type];
}

export function searchResultHref(result: SearchResult): string | undefined {
  switch (result.type) {
    case "Organization":
      return `/organizations/${result.id}`;
    case "Namespace":
      return `/namespaces/${result.id}`;
    case "Project":
      if (!result.namespace_id) {
        return undefined;
      }
      return `/namespaces/${result.namespace_id}/projects/${result.id}`;
    case "Issue":
      if (!result.namespace_id || !result.key) {
        return undefined;
      }
      return `/work/${result.namespace_id}/${result.key}`;
    case "Document":
      return `/documents/${result.id}`;
  }
}

export function groupSearchResults(
  items: readonly SearchResult[]
): { type: SearchResourceType; items: SearchResult[] }[] {
  const groups: { type: SearchResourceType; items: SearchResult[] }[] = [];
  for (const type of SEARCH_RESOURCE_TYPES) {
    const grouped = items.filter((item) => item.type === type);
    if (grouped.length > 0) {
      groups.push({ type, items: grouped });
    }
  }
  return groups;
}
