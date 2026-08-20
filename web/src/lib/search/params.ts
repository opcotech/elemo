import { z } from "zod";

import type { V1SearchGetData } from "@/lib/client";
import { SEARCH_RESOURCE_TYPES } from "@/lib/search/result";

export const SEARCH_PAGE_SIZE = 20;
export const SEARCH_PALETTE_PAGE_SIZE = 8;
export const SEARCH_DEBOUNCE_MS = 300;

export const searchRouteTypeSchema = z
  .enum(["all", ...SEARCH_RESOURCE_TYPES])
  .catch("all");

const optionalSearchId = z.preprocess(
  (value) => (typeof value === "string" && value.trim() ? value : undefined),
  z.string().max(160).optional()
);

export const searchRouteSearchSchema = z.object({
  q: z.string().max(200).catch(""),
  type: searchRouteTypeSchema,
  organization_id: optionalSearchId,
  namespace_id: optionalSearchId,
  project_id: optionalSearchId,
  page_token: z.string().max(2048).optional().catch(undefined),
});

export type SearchRouteSearch = z.infer<typeof searchRouteSearchSchema>;
export type SearchRouteType = SearchRouteSearch["type"];

export function hasActiveSearch(search: SearchRouteSearch): boolean {
  return (
    search.q.trim().length > 0 ||
    search.type !== "all" ||
    Boolean(search.organization_id) ||
    Boolean(search.namespace_id) ||
    Boolean(search.project_id)
  );
}

export function searchQueryFromRoute(
  search: SearchRouteSearch,
  pageSize = SEARCH_PAGE_SIZE
): NonNullable<V1SearchGetData["query"]> {
  const query: NonNullable<V1SearchGetData["query"]> = {
    page_size: pageSize,
  };
  const text = search.q.trim();
  if (text) {
    query.q = text;
  }
  if (search.type !== "all") {
    query.types = [search.type];
  }
  if (search.organization_id) {
    query.organization_id = search.organization_id;
  }
  if (search.namespace_id) {
    query.namespace_id = search.namespace_id;
  }
  if (search.project_id) {
    query.project_id = search.project_id;
  }
  if (search.page_token) {
    query.page_token = search.page_token;
  }
  return query;
}
