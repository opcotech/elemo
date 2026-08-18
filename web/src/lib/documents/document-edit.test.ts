import { describe, expect, it } from "vitest";

import {
  excerptFromContent,
  parseDocumentContent,
  parseDocumentExcerpt,
  parseDocumentTitle,
  resolveDocumentExcerpt,
} from "@/lib/documents/document-edit";

describe("document-edit helpers", () => {
  describe("parseDocumentTitle", () => {
    it("accepts a valid title", () => {
      expect(parseDocumentTitle("  Project Plan  ")).toEqual({
        ok: true,
        title: "Project Plan",
      });
    });

    it("rejects titles that are too short", () => {
      const result = parseDocumentTitle("ab");
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.error.length).toBeGreaterThan(0);
      }
    });

    it("rejects titles that are too long", () => {
      const result = parseDocumentTitle("a".repeat(121));
      expect(result.ok).toBe(false);
    });
  });

  describe("parseDocumentExcerpt", () => {
    it("clears blank excerpts", () => {
      expect(parseDocumentExcerpt("   ")).toEqual({
        ok: true,
        excerpt: null,
      });
    });

    it("accepts a valid excerpt", () => {
      expect(parseDocumentExcerpt("  Overview of the project plan  ")).toEqual({
        ok: true,
        excerpt: "Overview of the project plan",
      });
    });

    it("rejects excerpts that are too short", () => {
      const result = parseDocumentExcerpt("Too short");
      expect(result.ok).toBe(false);
    });

    it("rejects excerpts that are too long", () => {
      const result = parseDocumentExcerpt("a".repeat(501));
      expect(result.ok).toBe(false);
    });
  });

  describe("excerptFromContent", () => {
    it("returns the first 100 characters of collapsed content", () => {
      expect(excerptFromContent("a".repeat(150))).toBe("a".repeat(100));
    });

    it("returns the full body when it is shorter than 100 characters", () => {
      expect(excerptFromContent("  Overview of the project plan  ")).toBe(
        "Overview of the project plan"
      );
    });

    it("returns null when the body is too short for an excerpt", () => {
      expect(excerptFromContent("Too short")).toBeNull();
    });
  });

  describe("resolveDocumentExcerpt", () => {
    it("uses the first 100 characters of content when the excerpt is blank", () => {
      expect(resolveDocumentExcerpt("   ", "a".repeat(150))).toEqual({
        ok: true,
        excerpt: "a".repeat(100),
      });
    });

    it("keeps a filled excerpt", () => {
      expect(
        resolveDocumentExcerpt(
          "  Overview of the project plan  ",
          "Body that should not be used"
        )
      ).toEqual({
        ok: true,
        excerpt: "Overview of the project plan",
      });
    });
  });

  describe("parseDocumentContent", () => {
    it("accepts empty content", () => {
      expect(parseDocumentContent("   ")).toEqual({
        ok: true,
        content: "",
      });
    });

    it("accepts markdown of at least one character", () => {
      expect(parseDocumentContent("# Goals")).toEqual({
        ok: true,
        content: "# Goals",
      });
    });

    it("trims surrounding whitespace", () => {
      expect(parseDocumentContent("  Goals  ")).toEqual({
        ok: true,
        content: "Goals",
      });
    });
  });
});
