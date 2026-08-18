import { describe, expect, it } from "vitest";

import {
  documentCreateBody,
  documentCreateContextCopy,
  documentCreateFormSchema,
  documentCreateParentFromNavigation,
} from "./create";

describe("document create helpers", () => {
  it("sends only the title", () => {
    expect(
      documentCreateBody({
        title: "  Architecture  ",
      })
    ).toEqual({
      title: "Architecture",
    });
  });

  it("accepts a title-only form", () => {
    expect(
      documentCreateFormSchema.parse({
        title: "Architecture",
      })
    ).toEqual({
      title: "Architecture",
    });
  });

  it("resolves create parents from navigation context", () => {
    expect(
      documentCreateParentFromNavigation({
        type: "project",
        organizationId: "org-1",
        namespaceId: "ns-1",
        projectId: "proj-1",
      })
    ).toEqual({ type: "project", id: "proj-1" });
    expect(
      documentCreateParentFromNavigation({
        type: "namespace",
        organizationId: "org-1",
        namespaceId: "ns-1",
      })
    ).toEqual({ type: "namespace", id: "ns-1" });
    expect(
      documentCreateParentFromNavigation({
        type: "organization",
        organizationId: "org-1",
      })
    ).toEqual({ type: "organization", id: "org-1" });
    expect(
      documentCreateParentFromNavigation({
        type: "global",
        organizationId: "org-1",
      })
    ).toBeNull();
  });

  it("describes where a created document will live", () => {
    expect(
      documentCreateContextCopy({
        type: "project",
        namespaceName: "Product",
        projectName: "Platform",
      })
    ).toBe("Lives in Product. Related to Platform.");
    expect(
      documentCreateContextCopy({
        type: "namespace",
        namespaceName: "Product",
      })
    ).toBe("Lives in Product.");
    expect(
      documentCreateContextCopy({
        type: "organization",
        organizationName: "Acme",
      })
    ).toBe("Lives in Acme.");
    expect(documentCreateContextCopy({ type: "global" })).toBe(
      "Global context"
    );
  });
});
