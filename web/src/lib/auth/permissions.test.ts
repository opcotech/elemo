import { describe, expect, it } from "vitest";

import { Action, ResourceType, can, withResourceType } from "./permissions";

import { SYSTEM_NIL_ID } from "@/lib/utils";

describe("permission checks", () => {
  it("rejects missing and empty action sets", () => {
    expect(can(undefined, Action.OrganizationRead)).toBe(false);
    expect(can([], Action.OrganizationRead)).toBe(false);
    expect(can({ actions: [] }, Action.OrganizationRead)).toBe(false);
  });

  it("accepts exact action matches only", () => {
    expect(
      can({ actions: [Action.OrganizationDelete] }, Action.OrganizationDelete)
    ).toBe(true);
    expect(can([Action.OrganizationDelete], Action.OrganizationDelete)).toBe(
      true
    );
    expect(
      can({ actions: [Action.OrganizationRead] }, Action.OrganizationDelete)
    ).toBe(false);
  });

  it("does not treat * as allow-all", () => {
    expect(can({ actions: ["*"] }, Action.OrganizationDelete)).toBe(false);
    expect(can(["*"], Action.OrganizationRead)).toBe(false);
  });

  it("does not let write imply read", () => {
    const permissions = { actions: [Action.OrganizationUpdate] };

    expect(can(permissions, Action.OrganizationRead)).toBe(false);
    expect(can(permissions, Action.OrganizationDelete)).toBe(false);
    expect(can(permissions, Action.OrganizationUpdate)).toBe(true);
  });

  it("builds resource keys with explicit and system identifiers", () => {
    expect(withResourceType(ResourceType.Project, "project-1")).toBe(
      "Project:project-1"
    );
    expect(withResourceType(ResourceType.Installation)).toBe(
      `Installation:${SYSTEM_NIL_ID}`
    );
  });
});
