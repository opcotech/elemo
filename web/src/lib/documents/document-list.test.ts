import { describe, expect, it } from "vitest";

import type { PartialDocument } from "@/lib/api/types";
import {
  documentAuthorName,
  documentCreators,
  documentExcerpt,
  documentUpdatedAt,
  filterDocuments,
  isDocumentListSort,
  sortDocuments,
  visibleDocuments,
} from "@/lib/documents/document-list";

function document(
  overrides: Partial<PartialDocument> & Pick<PartialDocument, "id" | "title">
): PartialDocument {
  return {
    created_by: {
      id: "user-1",
      first_name: "Ada",
      last_name: "Lovelace",
      picture: null,
    },
    excerpt: null,
    created_at: null,
    updated_at: null,
    ...overrides,
  };
}

describe("document list helpers", () => {
  const alpha = document({
    id: "a",
    title: "Alpha plan",
    excerpt: "Overview of alpha",
    created_at: "2026-01-02T00:00:00Z",
    updated_at: "2026-04-01T00:00:00Z",
  });
  const beta = document({
    id: "b",
    title: "Beta notes",
    excerpt: "Later notes",
    created_by: {
      id: "user-2",
      first_name: "Grace",
      last_name: "Hopper",
      picture: null,
    },
    created_at: "2026-03-01T00:00:00Z",
  });
  const gamma = document({
    id: "c",
    title: "Gamma",
    created_at: "2026-02-01T00:00:00Z",
    updated_at: "2026-02-15T00:00:00Z",
  });

  it("filters by title, excerpt, or author", () => {
    expect(filterDocuments([alpha, beta, gamma], "alpha")).toEqual([alpha]);
    expect(filterDocuments([alpha, beta, gamma], "notes")).toEqual([beta]);
    expect(filterDocuments([alpha, beta, gamma], "lovelace")).toEqual([
      alpha,
      gamma,
    ]);
  });

  it("filters by creator", () => {
    expect(filterDocuments([alpha, beta, gamma], "", "user-2")).toEqual([beta]);
    expect(filterDocuments([alpha, beta, gamma], "", "all")).toEqual([
      alpha,
      beta,
      gamma,
    ]);
  });

  it("returns all documents when the query is blank", () => {
    expect(filterDocuments([alpha, beta], "  ")).toEqual([alpha, beta]);
  });

  it("sorts by last update descending", () => {
    expect(
      sortDocuments([alpha, beta, gamma], "updated-desc").map((item) => item.id)
    ).toEqual(["a", "b", "c"]);
  });

  it("sorts by created date descending", () => {
    expect(
      sortDocuments([alpha, beta, gamma], "created-desc").map((item) => item.id)
    ).toEqual(["b", "c", "a"]);
  });

  it("sorts by created date ascending", () => {
    expect(
      sortDocuments([alpha, beta, gamma], "created-asc").map((item) => item.id)
    ).toEqual(["a", "c", "b"]);
  });

  it("sorts by title", () => {
    expect(
      sortDocuments([gamma, beta, alpha], "title-asc").map((item) => item.id)
    ).toEqual(["a", "b", "c"]);
  });

  it("filters then sorts visible documents", () => {
    expect(
      visibleDocuments([beta, alpha, gamma], "plan", "title-asc").map(
        (item) => item.id
      )
    ).toEqual(["a"]);
    expect(
      visibleDocuments([beta, alpha, gamma], "", "updated-desc", "user-1").map(
        (item) => item.id
      )
    ).toEqual(["a", "c"]);
  });

  it("lists unique creators by name", () => {
    const withPicture = document({
      id: "d",
      title: "Delta",
      created_by: {
        id: "user-1",
        first_name: "Ada",
        last_name: "Lovelace",
        picture: "https://cdn.example/ada.png",
      },
    });
    expect(documentCreators([beta, withPicture, gamma])).toEqual([
      {
        id: "user-1",
        name: "Ada Lovelace",
        picture: "https://cdn.example/ada.png",
        initials: "AL",
      },
      {
        id: "user-2",
        name: "Grace Hopper",
        picture: null,
        initials: "GH",
      },
    ]);
  });

  it("uses the author name and omits blank excerpts", () => {
    expect(documentAuthorName(alpha)).toBe("Ada Lovelace");
    expect(documentExcerpt(alpha)).toBe("Overview of alpha");
    expect(documentExcerpt(gamma)).toBeNull();
    expect(
      documentExcerpt(document({ id: "d", title: "Delta", excerpt: "  " }))
    ).toBeNull();
  });

  it("falls back to created_at when updated_at is missing", () => {
    expect(documentUpdatedAt(alpha)).toBe("2026-04-01T00:00:00Z");
    expect(documentUpdatedAt(beta)).toBe("2026-03-01T00:00:00Z");
  });

  it("recognizes document list sorts", () => {
    expect(isDocumentListSort("updated-desc")).toBe(true);
    expect(isDocumentListSort("created-desc")).toBe(true);
    expect(isDocumentListSort("unknown")).toBe(false);
  });
});
