import type { QueryClient } from "@tanstack/react-query";

import { ApiError } from "@/lib/api/errors";
import { v1PermissionResourceGetOptions } from "@/lib/api/query-options";
import type { EffectiveActions } from "@/lib/api/types";
import {
  Action,
  ResourceType,
  can,
  withResourceType,
} from "@/lib/auth/permissions";

export async function loadResourcePermissions(
  queryClient: QueryClient,
  resourceType: ResourceType,
  resourceId?: string
): Promise<EffectiveActions> {
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

function defaultReadAction(resourceType: ResourceType): string {
  switch (resourceType) {
    case ResourceType.Namespace:
      return Action.NamespaceRead;
    case ResourceType.Project:
      return Action.ProjectRead;
    case ResourceType.Document:
      return Action.DocumentRead;
    case ResourceType.Issue:
      return Action.IssueRead;
    default:
      return Action.OrganizationRead;
  }
}

export async function requireResourcePermission(
  queryClient: QueryClient,
  resourceType: ResourceType,
  resourceId: string,
  action: string = defaultReadAction(resourceType)
): Promise<EffectiveActions> {
  const permissions = await loadResourcePermissions(
    queryClient,
    resourceType,
    resourceId
  );

  if (!can(permissions, action)) {
    throw new ApiError(403, "Permission denied");
  }

  return permissions;
}
