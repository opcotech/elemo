import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";

import { withResourceType } from "@/hooks/use-permissions";
import type { ResourceType } from "@/hooks/use-permissions";
import { v1PermissionResourceGetOptions } from "@/lib/api";
import type { PermissionKind } from "@/lib/api";
import { can } from "@/lib/auth/permissions";

export interface UseRequirePermissionOptions {
  resourceType: ResourceType;
  permissionKind: PermissionKind;
  resourceId?: string | (() => string | undefined);
}

/**
 * Hook that checks permissions immediately on mount and redirects if denied.
 * This ensures permission checks happen before any data fetching or loading states.
 *
 * @param options - Permission requirement options
 * @returns Object with isLoading state (true while checking permissions)
 */
export function useRequirePermission({
  resourceType,
  permissionKind,
  resourceId,
}: UseRequirePermissionOptions) {
  const navigate = useNavigate();

  // Resolve resource ID
  const resolvedResourceId =
    typeof resourceId === "function" ? resourceId() : resourceId;

  // Construct the resource ID string for permission checking
  const permissionResourceId = withResourceType(
    resourceType,
    resolvedResourceId
  );

  const {
    data: permissions,
    isLoading,
    error,
  } = useQuery(
    v1PermissionResourceGetOptions({
      path: {
        resourceId: permissionResourceId,
      },
    })
  );

  // Redirect immediately if permission check fails
  useEffect(() => {
    if (isLoading) return;

    // Redirect if there's an error or no permissions
    if (error || !permissions || permissions.length === 0) {
      navigate({ to: "/permission-denied" });
      return;
    }

    // Redirect if user doesn't have the required permission
    if (!can(permissions, permissionKind)) {
      navigate({ to: "/permission-denied" });
      return;
    }
  }, [isLoading, error, permissions, permissionKind, navigate]);

  return { isLoading };
}
