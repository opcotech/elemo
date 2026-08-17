import type { Query } from "@tanstack/react-query";
import { describe, expect, it } from "vitest";

import {
  isDocumentListQuery,
  isFolderListQuery,
} from "@/lib/documents/document-queries";

function queryWithId(id: string): Query {
  return {
    queryKey: [{ _id: id, path: { id: "parent-1" } }],
  } as unknown as Query;
}

describe("isDocumentListQuery", () => {
  it("matches parent document list operations", () => {
    expect(isDocumentListQuery(queryWithId("v1NamespacesDocumentsGet"))).toBe(
      true
    );
    expect(isDocumentListQuery(queryWithId("v1ProjectsDocumentsGet"))).toBe(
      true
    );
    expect(
      isDocumentListQuery(queryWithId("v1OrganizationsDocumentsGet"))
    ).toBe(true);
    expect(isDocumentListQuery(queryWithId("v1IssuesDocumentsGet"))).toBe(true);
  });

  it("matches folder list operations", () => {
    expect(isFolderListQuery(queryWithId("v1NamespacesFoldersGet"))).toBe(true);
    expect(isFolderListQuery(queryWithId("v1OrganizationsFoldersGet"))).toBe(
      true
    );
    expect(isFolderListQuery(queryWithId("v1NamespacesDocumentsGet"))).toBe(
      false
    );
  });

  it("does not match document detail or unrelated queries", () => {
    expect(isDocumentListQuery(queryWithId("v1DocumentGet"))).toBe(false);
    expect(isDocumentListQuery(queryWithId("v1NamespacesIssuesGet"))).toBe(
      false
    );
    expect(
      isDocumentListQuery({ queryKey: ["plain"] } as unknown as Query)
    ).toBe(false);
  });
});
