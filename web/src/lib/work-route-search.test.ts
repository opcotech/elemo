import { describe, expect, it } from "vitest";

import {
  resolveWorkScope,
  serializeWorkScope,
  workRouteSearchSchema,
  workScopeOptions,
} from "./work-route-search";

describe("workRouteSearchSchema", () => {
  it("applies stable projection defaults", () => {
    expect(workRouteSearchSchema.parse({})).toEqual({
      display: "comfortable",
      group: "status",
      layout: "board",
      sort: "rank:asc",
    });
  });

  it("preserves shared query state while switching projections", () => {
    const sharedState = {
      view: "view-web-delivery",
      scope: "project:project-web",
      filter: "urgent navigation",
      group: "assignee",
      sort: "priority:desc",
      display: "compact",
      selected: "work:lmo-101",
    } as const;

    for (const layout of ["board", "list", "table", "timeline"] as const) {
      expect(workRouteSearchSchema.parse({ ...sharedState, layout })).toEqual({
        ...sharedState,
        layout,
      });
    }
  });

  it("recovers invalid projection controls and rejects invalid selections", () => {
    expect(
      workRouteSearchSchema.parse({
        display: "tiny",
        group: "team",
        layout: "grid",
        sort: 42,
      })
    ).toMatchObject({
      display: "comfortable",
      group: "status",
      layout: "board",
      sort: "rank:asc",
    });

    expect(() =>
      workRouteSearchSchema.parse({ selected: "document:document-1" })
    ).toThrow();
    expect(() =>
      workRouteSearchSchema.parse({ scope: "https://example.com" })
    ).toThrow();
  });

  it("keeps only work selections and coerces empty selected values", () => {
    expect(
      workRouteSearchSchema.parse({
        selected: "work:ops-301",
        filter: "incident",
      })
    ).toMatchObject({
      selected: "work:ops-301",
      filter: "incident",
    });
  });

  it("limits URL scopes to the current route hierarchy", () => {
    const projectScope = {
      type: "project",
      namespaceId: "namespace-product",
      projectId: "project-web",
    } as const;

    expect(workScopeOptions(projectScope).map(serializeWorkScope)).toEqual([
      "project:project-web",
      "namespace:namespace-product",
      "global",
    ]);
    expect(
      resolveWorkScope("namespace:namespace-product", projectScope)
    ).toEqual({
      type: "namespace",
      namespaceId: "namespace-product",
    });
    expect(resolveWorkScope("project:another-project", projectScope)).toBe(
      projectScope
    );
  });

  it("serializes person and global scopes for URL ownership", () => {
    expect(serializeWorkScope({ type: "global" })).toBe("global");
    expect(serializeWorkScope({ type: "person", personId: "person-ada" })).toBe(
      "person:person-ada"
    );
    expect(
      resolveWorkScope("global", {
        type: "namespace",
        namespaceId: "namespace-product",
      })
    ).toEqual({ type: "global" });
  });
});
