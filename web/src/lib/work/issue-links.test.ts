import { describe, expect, it } from "vitest";

import {
  issueLinkFaviconSrc,
  issueLinkHostname,
  parseIssueLink,
  parseIssueLinkUrl,
} from "./issue-links";

describe("issue link helpers", () => {
  it("accepts http(s) URLs and normalizes a missing protocol", () => {
    expect(parseIssueLinkUrl("https://example.com/path")).toEqual({
      ok: true,
      url: "https://example.com/path",
    });
    expect(parseIssueLinkUrl("example.com")).toEqual({
      ok: true,
      url: "https://example.com/",
    });
  });

  it("rejects empty and non-http values", () => {
    expect(parseIssueLinkUrl("   ").ok).toBe(false);
    expect(parseIssueLinkUrl("javascript:alert(1)").ok).toBe(false);
    expect(parseIssueLinkUrl("not a url").ok).toBe(false);
  });

  it("requires a visible label", () => {
    expect(parseIssueLink("https://example.com", "  Spec  ")).toEqual({
      ok: true,
      url: "https://example.com/",
      label: "Spec",
    });
    expect(parseIssueLink("https://example.com", "   ").ok).toBe(false);
    expect(parseIssueLink("https://example.com", "x".repeat(121)).ok).toBe(
      false
    );
  });

  it("extracts hostname and builds the favicon URL", () => {
    expect(issueLinkHostname("https://docs.example.com/guide")).toBe(
      "docs.example.com"
    );
    expect(issueLinkHostname("not-a-url")).toBeNull();
    expect(issueLinkFaviconSrc("docs.example.com")).toBe(
      "https://www.google.com/s2/favicons?domain=docs.example.com&sz=32"
    );
  });
});
