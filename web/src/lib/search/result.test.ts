import { describe, expect, it } from "vitest";

import type { SearchResult } from "@/lib/api/types";
import {
  groupSearchResults,
  searchResultEntityType,
  searchResultHref,
} from "@/lib/search/result";

function result(
  overrides: Pick<SearchResult, "id" | "type" | "title"> & Partial<SearchResult>
): SearchResult {
  return {
    created_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("searchResultHref", () => {
  it("maps organizations, namespaces, and documents from slugs", () => {
    expect(
      searchResultHref(
        result({
          id: "org-1",
          type: "Organization",
          title: "Acme",
          organization_slug: "acme",
        })
      )
    ).toBe("/organizations/acme");
    expect(
      searchResultHref(
        result({
          id: "ns-1",
          type: "Namespace",
          title: "Product",
          organization_slug: "acme",
          namespace_slug: "product",
        })
      )
    ).toBe("/organizations/acme/namespaces/product");
    expect(
      searchResultHref(result({ id: "doc-1", type: "Document", title: "Spec" }))
    ).toBe("/documents/doc-1");
  });

  it("requires slug ancestry for projects and issues, not xids", () => {
    expect(
      searchResultHref(
        result({ id: "proj-1", type: "Project", title: "Platform" })
      )
    ).toBeUndefined();
    expect(
      searchResultHref(
        result({
          id: "proj-1",
          type: "Project",
          title: "Platform",
          namespace_id: "ns-1",
          key: "PLAT",
          organization_slug: "acme",
          namespace_slug: "product",
        })
      )
    ).toBe("/organizations/acme/namespaces/product/projects/PLAT");
    expect(
      searchResultHref(
        result({
          id: "issue-1",
          type: "Issue",
          title: "Broken search",
          key: "LMO-12",
        })
      )
    ).toBeUndefined();
    expect(
      searchResultHref(
        result({
          id: "issue-1",
          type: "Issue",
          title: "Broken search",
          key: "LMO-12",
          namespace_id: "ns-1",
          organization_slug: "acme",
          namespace_slug: "product",
        })
      )
    ).toBe("/work/acme/product/LMO-12");
  });

  it("maps issue hits to work-item entity types", () => {
    expect(
      searchResultEntityType(
        result({ id: "issue-1", type: "Issue", title: "Broken search" })
      )
    ).toBe("work-item");
  });
});

describe("groupSearchResults", () => {
  it("groups hits in resource-type order", () => {
    const groups = groupSearchResults([
      result({ id: "issue-1", type: "Issue", title: "A" }),
      result({ id: "org-1", type: "Organization", title: "B" }),
      result({ id: "issue-2", type: "Issue", title: "C" }),
    ]);

    expect(groups.map((group) => group.type)).toEqual([
      "Organization",
      "Issue",
    ]);
    expect(groups[1]?.items).toHaveLength(2);
  });
});
