import { describe, expect, it } from "vitest";

import { resolveNavigationContext } from "./use-navigation-context";

describe("resolveNavigationContext", () => {
  it("treats organization routes as organization context", () => {
    expect(
      resolveNavigationContext("/organizations/acme", {
        organizationSlug: "acme",
        organizationId: "org-xid",
      })
    ).toEqual({
      type: "organization",
      organizationSlug: "acme",
      organizationId: "org-xid",
    });
    expect(
      resolveNavigationContext("/organizations/acme/documents", {
        organizationSlug: "acme",
        organizationId: "org-xid",
      })
    ).toEqual({
      type: "organization",
      organizationSlug: "acme",
      organizationId: "org-xid",
    });
  });

  it("keeps settings admin routes global even when slugs and ids are present", () => {
    expect(
      resolveNavigationContext("/settings/organizations/acme", {
        organizationSlug: "acme",
        namespaceSlug: "platform",
        projectKey: "PLAT",
        organizationId: "org-xid",
        namespaceId: "ns-xid",
        projectId: "proj-xid",
      })
    ).toEqual({ type: "global" });
  });

  it("prefers project, then namespace, over organization", () => {
    expect(
      resolveNavigationContext(
        "/organizations/acme/namespaces/platform/projects/PLAT",
        {
          organizationSlug: "acme",
          namespaceSlug: "platform",
          projectKey: "PLAT",
          organizationId: "org-xid",
          namespaceId: "ns-xid",
          projectId: "proj-xid",
        }
      )
    ).toEqual({
      type: "project",
      organizationSlug: "acme",
      namespaceSlug: "platform",
      projectKey: "PLAT",
      organizationId: "org-xid",
      namespaceId: "ns-xid",
      projectId: "proj-xid",
    });
    expect(
      resolveNavigationContext("/organizations/acme/namespaces/platform", {
        organizationSlug: "acme",
        namespaceSlug: "platform",
        organizationId: "org-xid",
        namespaceId: "ns-xid",
      })
    ).toEqual({
      type: "namespace",
      organizationSlug: "acme",
      namespaceSlug: "platform",
      organizationId: "org-xid",
      namespaceId: "ns-xid",
    });
  });

  it("keeps URL identity distinct from resolved xids", () => {
    const context = resolveNavigationContext("/organizations/acme", {
      organizationSlug: "acme",
      organizationId: "org-xid",
    });
    expect(context.organizationSlug).toBe("acme");
    expect(context.organizationId).toBe("org-xid");
    expect(context.organizationSlug).not.toBe(context.organizationId);
  });

  it("returns global context when no operational parent is present", () => {
    expect(resolveNavigationContext("/", {})).toEqual({ type: "global" });
    expect(resolveNavigationContext("/documents", {})).toEqual({
      type: "global",
    });
  });
});
