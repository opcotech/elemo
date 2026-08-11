import { describe, expect, it } from "vitest";

import { ResourceType, can, withResourceType } from "./permissions";

import type { Permission, PermissionKind } from "@/lib/api/types";
import { SYSTEM_NIL_ID } from "@/lib/utils";

function permission(kind: PermissionKind): Permission {
  return { kind } as Permission;
}

describe("permission checks", () => {
  it("rejects missing and empty permission sets", () => {
    expect(can(undefined, "read")).toBe(false);
    expect(can([], "read")).toBe(false);
  });

  it("accepts exact and wildcard grants", () => {
    expect(can([permission("delete")], "delete")).toBe(true);
    expect(can([permission("*")], "delete")).toBe(true);
  });

  it("lets write imply read without implying other actions", () => {
    const permissions = [permission("write")];

    expect(can(permissions, "read")).toBe(true);
    expect(can(permissions, "delete")).toBe(false);
    expect(can([permission("read")], "write")).toBe(false);
  });

  it("builds resource keys with explicit and system identifiers", () => {
    expect(withResourceType(ResourceType.Project, "project-1")).toBe(
      "Project:project-1"
    );
    expect(withResourceType(ResourceType.Organization)).toBe(
      `Organization:${SYSTEM_NIL_ID}`
    );
  });
});
