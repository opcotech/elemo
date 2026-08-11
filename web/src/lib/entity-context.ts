import type { QueryClient } from "@tanstack/react-query";

import { ApiError } from "@/lib/api/errors";
import { v1PermissionResourceGetOptions } from "@/lib/api/query-options";
import type { Permission, PermissionKind } from "@/lib/api/types";
import { can, withResourceType } from "@/lib/auth/permissions";
import type { ResourceType } from "@/lib/auth/permissions";

export async function loadResourcePermissions(
  queryClient: QueryClient,
  resourceType: ResourceType,
  resourceId?: string
): Promise<Permission[]> {
  // fetchQuery respects invalidation; ensureQueryData would keep serving a
  // stale permission snapshot after grant/revoke mutations.
  return queryClient.fetchQuery(
    v1PermissionResourceGetOptions({
      path: {
        resourceId: withResourceType(resourceType, resourceId),
      },
    })
  );
}

export async function requireResourcePermission(
  queryClient: QueryClient,
  resourceType: ResourceType,
  resourceId: string,
  permissionKind: PermissionKind = "read"
): Promise<Permission[]> {
  const permissions = await loadResourcePermissions(
    queryClient,
    resourceType,
    resourceId
  );

  if (!can(permissions, permissionKind)) {
    throw new ApiError(403, "Permission denied");
  }

  return permissions;
}
