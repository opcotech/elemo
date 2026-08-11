import { describe, expect, it } from "vitest";

import { resolveDemoPerson } from "./demo-person";
import { mockPeople } from "./fixtures";

describe("resolveDemoPerson", () => {
  it("prefers a case-insensitive email match", () => {
    expect(
      resolveDemoPerson({
        email: "GRACE@EXAMPLE.TEST",
        username: "ada",
      }).id
    ).toBe("person-grace");
  });

  it("falls back to the authenticated username", () => {
    expect(resolveDemoPerson({ username: "@katherine" }).id).toBe(
      "person-katherine"
    );
  });

  it("matches usernames without a leading @", () => {
    expect(resolveDemoPerson({ username: "ada" }).id).toBe("person-ada");
  });

  it("uses the documented first fixture for an unmatched identity", () => {
    expect(resolveDemoPerson({ email: "real@example.com" })).toBe(
      mockPeople[0]
    );
    expect(resolveDemoPerson(null)).toBe(mockPeople[0]);
    expect(resolveDemoPerson(undefined)).toBe(mockPeople[0]);
  });
});
