import { describe, expect, it } from "vitest";

import {
  identityFromLoaderData,
  identityFromMatches,
  requireIssueKey,
  requireNamespaceSlug,
  requireOrganizationSlug,
  requireProjectKey,
} from "./route-identity";

describe("route identifier validation", () => {
  it("accepts canonical slug and key segments", () => {
    expect(requireOrganizationSlug("acme")).toBe("acme");
    expect(requireNamespaceSlug("platform")).toBe("platform");
    expect(requireProjectKey("PLAT")).toBe("PLAT");
    expect(requireIssueKey("PLAT-1")).toBe("PLAT-1");
  });

  it("rejects xid, reserved, and noncanonical segments without fallback", () => {
    const xid = "9bsv0s46s6s002p9ltq0";
    expect(() => requireOrganizationSlug(xid)).toThrow();
    expect(() => requireOrganizationSlug("new")).toThrow();
    expect(() => requireOrganizationSlug("Acme")).toThrow();
    expect(() => requireNamespaceSlug("new")).toThrow();
    expect(() => requireProjectKey("plat")).toThrow();
    expect(() => requireProjectKey("NEW")).toThrow();
    expect(() => requireIssueKey("plat-1")).toThrow();
  });
});

describe("identityFromLoaderData", () => {
  it("reads resolved xids from nested loader entities, not URL params", () => {
    expect(
      identityFromLoaderData({
        organization: { id: "org-xid", slug: "acme" },
        namespace: { id: "ns-xid", slug: "platform" },
        project: { id: "proj-xid", key: "PLAT" },
      })
    ).toEqual({
      organizationId: "org-xid",
      namespaceId: "ns-xid",
      projectId: "proj-xid",
    });
  });

  it("merges identity across parent route matches", () => {
    expect(
      identityFromMatches([
        {
          loaderData: {
            organization: { id: "org-xid", slug: "acme" },
          },
        },
        {
          loaderData: {
            namespace: { id: "ns-xid", slug: "platform" },
          },
        },
      ])
    ).toEqual({
      organizationId: "org-xid",
      namespaceId: "ns-xid",
      projectId: undefined,
    });
  });
});
