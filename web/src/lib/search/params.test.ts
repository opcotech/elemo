import { describe, expect, it } from "vitest";

import {
  SEARCH_PAGE_SIZE,
  hasActiveSearch,
  searchQueryFromRoute,
  searchRouteSearchSchema,
} from "@/lib/search/params";

describe("searchRouteSearchSchema", () => {
  it("defaults to an empty global search", () => {
    expect(searchRouteSearchSchema.parse({})).toEqual({
      q: "",
      type: "all",
    });
  });

  it("preserves API filters and cursor pagination", () => {
    expect(
      searchRouteSearchSchema.parse({
        q: "projection",
        type: "Issue",
        organization_id: "org-1",
        namespace_id: "ns-1",
        project_id: "proj-1",
        page_token: "cursor-2",
      })
    ).toEqual({
      q: "projection",
      type: "Issue",
      organization_id: "org-1",
      namespace_id: "ns-1",
      project_id: "proj-1",
      page_token: "cursor-2",
    });
  });

  it("coerces invalid legacy filters and pagination", () => {
    expect(
      searchRouteSearchSchema.parse({
        q: 10,
        type: "work-item",
        scope: "namespace:namespace-product",
        selected: "work:ops-301",
        page: "3",
        organization_id: "",
      })
    ).toEqual({
      q: "",
      type: "all",
    });
  });
});

describe("searchQueryFromRoute", () => {
  it("omits empty text and all-types filters", () => {
    expect(
      searchQueryFromRoute({
        q: "  ",
        type: "all",
      })
    ).toEqual({
      page_size: SEARCH_PAGE_SIZE,
    });
  });

  it("sends type, scope, and cursor fields to the API", () => {
    expect(
      searchQueryFromRoute({
        q: "  search index  ",
        type: "Document",
        organization_id: "org-1",
        namespace_id: "ns-1",
        page_token: "cursor-2",
      })
    ).toEqual({
      q: "search index",
      types: ["Document"],
      organization_id: "org-1",
      namespace_id: "ns-1",
      page_size: SEARCH_PAGE_SIZE,
      page_token: "cursor-2",
    });
  });
});

describe("hasActiveSearch", () => {
  it("treats filters without a query as an active search", () => {
    expect(hasActiveSearch({ q: "", type: "all" })).toBe(false);
    expect(hasActiveSearch({ q: "ops", type: "all" })).toBe(true);
    expect(hasActiveSearch({ q: "", type: "all", namespace_id: "ns-1" })).toBe(
      true
    );
    expect(hasActiveSearch({ q: "", type: "Issue" })).toBe(true);
  });
});
