import type { SearchResult } from "@/lib/api/types";
import type { AppEntityType } from "@/lib/entity-types";
import {
  documentPath,
  namespacePath,
  organizationPath,
  projectPath,
  workItemPath,
} from "@/lib/paths";

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
      if (!result.organization_slug) {
        return undefined;
      }
      return organizationPath({ organizationSlug: result.organization_slug });
    case "Namespace":
      if (!result.organization_slug || !result.namespace_slug) {
        return undefined;
      }
      return namespacePath({
        organizationSlug: result.organization_slug,
        namespaceSlug: result.namespace_slug,
      });
    case "Project":
      if (!result.organization_slug || !result.namespace_slug || !result.key) {
        return undefined;
      }
      return projectPath({
        organizationSlug: result.organization_slug,
        namespaceSlug: result.namespace_slug,
        projectKey: result.key,
      });
    case "Issue":
      if (!result.organization_slug || !result.namespace_slug || !result.key) {
        return undefined;
      }
      return workItemPath({
        organizationSlug: result.organization_slug,
        namespaceSlug: result.namespace_slug,
        issueKey: result.key,
      });
    case "Document":
      return documentPath(result.id);
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
