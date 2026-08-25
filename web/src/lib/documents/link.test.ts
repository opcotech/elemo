import { describe, expect, it } from "vitest";

import {
  filterAvailableDocuments,
  relatedDocumentCatalogQueryOptions,
} from "./link";

describe("document link helpers", () => {
  it("filters already related documents from the picker", () => {
    const documents = [{ id: "doc-1" }, { id: "doc-2" }, { id: "doc-3" }];

    expect(
      filterAvailableDocuments(documents, new Set(["doc-2"])).map(
        (document) => document.id
      )
    ).toEqual(["doc-1", "doc-3"]);
  });

  it("loads the picker catalog from namespace documents", () => {
    const options = relatedDocumentCatalogQueryOptions("org-1", "ns-1");
    const queryKey = JSON.stringify(options.queryKey);

    expect(queryKey).toContain("v1NamespacesDocumentsGet");
    expect(queryKey).toContain("ns-1");
    expect(queryKey).toContain('"all":true');
  });
});
