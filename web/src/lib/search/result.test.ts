import { describe, expect, it } from "vitest";

import type { SearchResult } from "@/lib/client";
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
  it("maps organizations, namespaces, and documents by id", () => {
    expect(
      searchResultHref(
        result({ id: "org-1", type: "Organization", title: "Acme" })
      )
    ).toBe("/organizations/org-1");
    expect(
      searchResultHref(
        result({ id: "ns-1", type: "Namespace", title: "Product" })
      )
    ).toBe("/namespaces/ns-1");
    expect(
      searchResultHref(result({ id: "doc-1", type: "Document", title: "Spec" }))
    ).toBe("/documents/doc-1");
  });

  it("requires namespace ancestry for projects and issues", () => {
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
        })
      )
    ).toBe("/namespaces/ns-1/projects/proj-1");
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
        })
      )
    ).toBe("/work/ns-1/LMO-12");
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
