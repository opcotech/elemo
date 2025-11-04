import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useCallback, useState } from "react";

import type { PendingPermission } from "@/components/organizations/role-permission-draft";
import {
  v1OrganizationRolePermissionAddMutation,
  v1OrganizationRolePermissionsGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { showErrorToast } from "@/lib/toast";

interface UsePendingPermissionsProps {
  organizationId: string;
}

export function usePendingPermissions({
  organizationId,
}: UsePendingPermissionsProps) {
  const queryClient = useQueryClient();
  const [pendingPermissions, setPendingPermissions] = useState<
    PendingPermission[]
  >([]);
  const [isCreatingPermissions, setIsCreatingPermissions] = useState(false);

  const permissionMutation = useMutation(
    v1OrganizationRolePermissionAddMutation()
  );

  const addPermission = useCallback((permission: PendingPermission) => {
    setPendingPermissions((prev) => [...prev, permission]);
  }, []);

  const removePermission = useCallback((index: number) => {
    setPendingPermissions((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const clearPermissions = useCallback(() => {
    setPendingPermissions([]);
  }, []);

  const createPermissions = useCallback(
    async (targetRoleId: string) => {
      if (pendingPermissions.length === 0) {
        return { success: 0, failed: 0, total: 0 };
      }

      setIsCreatingPermissions(true);
      const errors: string[] = [];
      let successCount = 0;

      try {
        // Create permissions sequentially to avoid overwhelming the API
        for (const permission of pendingPermissions) {
          try {
            await permissionMutation.mutateAsync({
              path: {
                id: organizationId,
                role_id: targetRoleId,
              },
              body: {
                target: permission.target,
                kind: permission.kind as
                  | "read"
                  | "write"
                  | "create"
                  | "delete"
                  | "*",
              },
            });
            successCount++;
          } catch (error) {
            const errorMessage =
              error instanceof Error
                ? error.message
                : "Failed to create permission";
            errors.push(
              `${permission.resourceType}:${permission.resourceId} (${permission.kind}) - ${errorMessage}`
            );
          }
        }

        // Show error toast if any permissions failed
        if (errors.length > 0) {
          showErrorToast(
            `${errors.length} permission(s) failed`,
            errors.join("; ")
          );
        }

        // Invalidate permissions query to refresh the list
        if (successCount > 0) {
          queryClient.invalidateQueries({
            queryKey: v1OrganizationRolePermissionsGetOptions({
              path: {
                id: organizationId,
                role_id: targetRoleId,
              },
            }).queryKey,
          });
        }

        return {
          success: successCount,
          failed: errors.length,
          total: pendingPermissions.length,
        };
      } finally {
        setIsCreatingPermissions(false);
      }
    },
    [organizationId, pendingPermissions, permissionMutation, queryClient]
  );

  return {
    pendingPermissions,
    addPermission,
    removePermission,
    clearPermissions,
    createPermissions,
    isCreatingPermissions,
    hasPendingPermissions: pendingPermissions.length > 0,
  };
}
