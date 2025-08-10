import type { Permission, PermissionKind } from "@/lib/api";

/**
 * Checks if the given permissions allow the specified action.
 *
 * @param permissions - Array of permissions to check
 * @param kind - The permission kind to check for
 * @returns true if the action is allowed, false otherwise
 */
export function can(
  permissions: Permission[] | undefined,
  kind: PermissionKind
) {
  if (!permissions || permissions.length === 0) {
    return false;
  }

  const permissionKinds = new Set(permissions.map((p) => p.kind));

  if (permissionKinds.has("*") || permissionKinds.has(kind)) {
    return true;
  }

  // Special case: write permission also grants access to read
  if (kind === "read" && permissionKinds.has("write")) {
    return true;
  }

  return false;
}
