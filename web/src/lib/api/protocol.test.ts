import { describe, expect, it } from "vitest";

import {
  buildUpstreamUrl,
  isPublicApiRequest,
  isSafeApiPath,
  transportRequestSchema,
} from "./protocol";

describe("API transport protocol", () => {
  it.each([
    "https://attacker.example/v1/users",
    "//attacker.example/v1/users",
    "/v1/../oauth/token",
    "/v1/users\\..\\oauth",
    "/v1/users#fragment",
  ])("rejects unsafe path %s", (path) => {
    expect(isSafeApiPath(path)).toBe(false);
  });

  it("accepts relative API paths with query strings", () => {
    expect(isSafeApiPath("/v1/users?limit=20&offset=0")).toBe(true);
  });

  it("builds upstream URLs without an /api prefix", () => {
    expect(
      buildUpstreamUrl("http://127.0.0.1:35478", "/v1/notifications").href
    ).toBe("http://127.0.0.1:35478/v1/notifications");
    expect(() =>
      buildUpstreamUrl("http://127.0.0.1:35478", "/api/v1/notifications")
    ).toThrow();
  });

  it("rejects forbidden and smuggled headers", () => {
    expect(() =>
      transportRequestSchema.parse({
        method: "GET",
        path: "/v1/users",
        headers: { authorization: "Bearer browser-token" },
      })
    ).toThrow();
    expect(() =>
      transportRequestSchema.parse({
        method: "GET",
        path: "/v1/users",
        headers: { accept: "application/json\r\nx-injected: true" },
      })
    ).toThrow();
  });

  it("allows only explicit public endpoints", () => {
    expect(
      isPublicApiRequest({
        method: "GET",
        path: "/v1/users/reset?email=a@b.co",
      })
    ).toBe(true);
    expect(
      isPublicApiRequest({
        method: "POST",
        path: "/v1/organizations/org-1/members/accept",
      })
    ).toBe(true);
    expect(
      isPublicApiRequest({ method: "GET", path: "/v1/organizations" })
    ).toBe(false);
  });
});
