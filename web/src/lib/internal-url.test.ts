import { describe, expect, it } from "vitest";

import { internalPath, isSafeInternalPath } from "./internal-url";

describe("safe internal paths", () => {
  it("accepts application routes with search and hash state", () => {
    expect(isSafeInternalPath("/namespaces/ns-1/work?layout=list#top")).toBe(
      true
    );
    expect(internalPath("/my-work?view=mine")).toBe("/my-work?view=mine");
  });

  it.each([
    "https://example.com",
    "//example.com/path",
    "/\\example.com",
    "/work/one two",
    "my-work",
    "",
  ])("rejects unsafe navigation target %s", (value) => {
    expect(isSafeInternalPath(value)).toBe(false);
  });
});
