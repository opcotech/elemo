import { describe, expect, it } from "vitest";

import { sanitizeRedirectTarget } from "./redirect";

describe("sanitizeRedirectTarget", () => {
  it("keeps same-origin paths and query strings", () => {
    expect(
      sanitizeRedirectTarget(
        "https://elemo.test/settings?tab=security",
        "https://elemo.test"
      )
    ).toBe("/settings?tab=security");
  });

  it.each([
    "https://attacker.example/phish",
    "//attacker.example/phish",
    "javascript:alert(1)",
    "/settings\\redirect",
    "/login",
  ])("falls back for unsafe redirect %s", (redirect) => {
    expect(sanitizeRedirectTarget(redirect, "https://elemo.test")).toBe("/");
  });
});
