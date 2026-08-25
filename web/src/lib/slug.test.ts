import { describe, expect, it } from "vitest";

import {
  isCanonicalIssueKey,
  isCanonicalNamespaceSlug,
  isCanonicalOrganizationSlug,
  isCanonicalProjectKey,
  isCanonicalSlug,
  isXidShaped,
  namespaceSlugMessage,
  organizationSlugMessage,
  suggestSlug,
} from "./slug";

const XID_SHAPED = "9bsv0s46s6s002p9ltq0";

describe("slug validation", () => {
  it("accepts canonical kebab-case", () => {
    expect(isCanonicalSlug("abc")).toBe(true);
    expect(isCanonicalSlug("acme-inc")).toBe(true);
    expect(isCanonicalSlug("org-2")).toBe(true);
    expect(isCanonicalSlug("a".repeat(3))).toBe(true);
    expect(isCanonicalSlug(`ab-${"c".repeat(47)}`.slice(0, 50))).toBe(true);
  });

  it("rejects boundaries, separators, and noncanonical input", () => {
    expect(isCanonicalSlug("")).toBe(false);
    expect(isCanonicalSlug("ab")).toBe(false);
    expect(isCanonicalSlug("a".repeat(51))).toBe(false);
    expect(isCanonicalSlug("Acme")).toBe(false);
    expect(isCanonicalSlug(" acme")).toBe(false);
    expect(isCanonicalSlug("acme ")).toBe(false);
    expect(isCanonicalSlug("acme_inc")).toBe(false);
    expect(isCanonicalSlug("-acme")).toBe(false);
    expect(isCanonicalSlug("acme-")).toBe(false);
    expect(isCanonicalSlug("acme--inc")).toBe(false);
    expect(isCanonicalSlug("acme%2Finc")).toBe(false);
    expect(isCanonicalSlug("ácme")).toBe(false);
    expect(isCanonicalSlug("аcme")).toBe(false);
    expect(isCanonicalSlug("acme/inc")).toBe(false);
    expect(isCanonicalSlug("acme.inc")).toBe(false);
  });

  it("rejects xid-shaped values", () => {
    expect(isXidShaped(XID_SHAPED)).toBe(true);
    expect(isCanonicalSlug(XID_SHAPED)).toBe(false);
  });

  it("rejects reserved organization and namespace slugs", () => {
    expect(isCanonicalOrganizationSlug("acme")).toBe(true);
    expect(isCanonicalOrganizationSlug("new")).toBe(false);
    expect(isCanonicalOrganizationSlug("join")).toBe(false);
    expect(isCanonicalNamespaceSlug("platform")).toBe(true);
    expect(isCanonicalNamespaceSlug("new")).toBe(false);
    expect(isCanonicalNamespaceSlug("join")).toBe(true);
  });
});

describe("project and issue keys", () => {
  it("accepts uppercase 2–6 letter project keys except NEW", () => {
    expect(isCanonicalProjectKey("AB")).toBe(true);
    expect(isCanonicalProjectKey("PLAT")).toBe(true);
    expect(isCanonicalProjectKey("ABCDEF")).toBe(true);
    expect(isCanonicalProjectKey("A")).toBe(false);
    expect(isCanonicalProjectKey("ABCDEFG")).toBe(false);
    expect(isCanonicalProjectKey("plat")).toBe(false);
    expect(isCanonicalProjectKey("PL-T")).toBe(false);
    expect(isCanonicalProjectKey("NEW")).toBe(false);
  });

  it("accepts composite issue keys", () => {
    expect(isCanonicalIssueKey("PLAT-1")).toBe(true);
    expect(isCanonicalIssueKey("AB-99")).toBe(true);
    expect(isCanonicalIssueKey("plat-1")).toBe(false);
    expect(isCanonicalIssueKey("PLAT")).toBe(false);
    expect(isCanonicalIssueKey("PLAT-")).toBe(false);
  });
});

describe("slug field messages", () => {
  it("explains empty, reserved, and xid-shaped values", () => {
    expect(organizationSlugMessage("")).toBe("Slug is required");
    expect(organizationSlugMessage("new")).toBe("This slug is reserved");
    expect(organizationSlugMessage("join")).toBe("This slug is reserved");
    expect(organizationSlugMessage(XID_SHAPED)).toBe(
      "Slug cannot look like an identifier"
    );
    expect(namespaceSlugMessage("new")).toBe("This slug is reserved");
    expect(namespaceSlugMessage("join")).toBeUndefined();
    expect(organizationSlugMessage("acme")).toBeUndefined();
  });
});

describe("suggestSlug", () => {
  it("keeps already-canonical names", () => {
    expect(suggestSlug("acme")).toBe("acme");
    expect(suggestSlug("nova-labs")).toBe("nova-labs");
  });

  it("derives kebab-case from display names", () => {
    expect(suggestSlug("Acme Inc")).toBe("acme-inc");
    expect(suggestSlug("Nova_Labs")).toBe("nova-labs");
    expect(suggestSlug("  Product Platform  ")).toBe("product-platform");
  });

  it("drops punctuation, unicode, and confusables", () => {
    expect(suggestSlug("Acme, Inc.")).toBe("acme-inc");
    expect(suggestSlug("ácme")).toBe("cme");
    expect(suggestSlug("Hello 🌍 World")).toBe("hello-world");
  });

  it("percent-decodes then canonicalizes", () => {
    expect(suggestSlug("Acme%20Inc")).toBe("acme-inc");
  });

  it("returns empty when the result would be invalid", () => {
    expect(suggestSlug("ab")).toBe("");
    expect(suggestSlug("---")).toBe("");
    expect(suggestSlug("")).toBe("");
    expect(suggestSlug(XID_SHAPED)).toBe("");
  });
});
