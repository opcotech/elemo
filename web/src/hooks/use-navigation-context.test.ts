import { describe, expect, it } from "vitest";

import { resolveNavigationContext } from "./use-navigation-context";

describe("resolveNavigationContext", () => {
  it("treats organization routes as organization context", () => {
    expect(
      resolveNavigationContext("/organizations/org-1", {
        organizationId: "org-1",
      })
    ).toEqual({
      type: "organization",
      organizationId: "org-1",
    });
    expect(
      resolveNavigationContext("/organizations/org-1/documents", {
        organizationId: "org-1",
      })
    ).toEqual({
      type: "organization",
      organizationId: "org-1",
    });
  });

  it("keeps settings admin routes global even when ids are present", () => {
    expect(
      resolveNavigationContext("/settings/organizations/org-1", {
        organizationId: "org-1",
        namespaceId: "ns-1",
        projectId: "proj-1",
      })
    ).toEqual({ type: "global" });
  });

  it("prefers project, then namespace, over organization", () => {
    expect(
      resolveNavigationContext("/namespaces/ns-1/projects/proj-1", {
        organizationId: "org-1",
        namespaceId: "ns-1",
        projectId: "proj-1",
      })
    ).toEqual({
      type: "project",
      organizationId: "org-1",
      namespaceId: "ns-1",
      projectId: "proj-1",
    });
    expect(
      resolveNavigationContext("/namespaces/ns-1", {
        organizationId: "org-1",
        namespaceId: "ns-1",
      })
    ).toEqual({
      type: "namespace",
      organizationId: "org-1",
      namespaceId: "ns-1",
    });
  });

  it("returns global context when no operational parent is present", () => {
    expect(resolveNavigationContext("/", {})).toEqual({ type: "global" });
    expect(resolveNavigationContext("/documents", {})).toEqual({
      type: "global",
    });
  });
});
