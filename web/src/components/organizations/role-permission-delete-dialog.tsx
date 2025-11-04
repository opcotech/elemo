import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Trash2 } from "lucide-react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import type { Permission } from "@/lib/api";
import {
  v1OrganizationRolePermissionRemoveMutation,
  v1OrganizationRolePermissionsGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { getFieldValue } from "@/lib/forms";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

interface RolePermissionDeleteDialogProps {
  permission: Permission;
  organizationId: string;
  roleId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function RolePermissionDeleteDialog({
  permission,
  organizationId,
  roleId,
  open,
  onOpenChange,
  onSuccess,
}: RolePermissionDeleteDialogProps) {
  const queryClient = useQueryClient();

  const mutation = useMutation(v1OrganizationRolePermissionRemoveMutation());

  const handleDelete = () => {
    mutation.mutate(
      {
        path: {
          id: organizationId,
          role_id: roleId,
          permission_id: permission.id,
        },
      },
      {
        onSuccess: () => {
          showSuccessToast(
            "Permission removed",
            "Permission removed successfully"
          );

          // Invalidate queries to refresh the permissions list
          queryClient.invalidateQueries({
            queryKey: v1OrganizationRolePermissionsGetOptions({
              path: {
                id: organizationId,
                role_id: roleId,
              },
            }).queryKey,
          });

          onSuccess?.();
          onOpenChange(false);
        },
        onError: (error) => {
          showErrorToast("Failed to remove permission", error.message);
        },
      }
    );
  };

  // Parse permission target to show in confirmation
  const parseTarget = (target: string): string => {
    const [resourceType, resourceId] = target.split(":");
    if (!resourceId) {
      return target;
    }
    const displayId =
      resourceId === "00000000000000000000"
        ? "System"
        : resourceId.slice(0, 8) + "...";
    return `${resourceType}: ${displayId}`;
  };

  const targetDisplay = parseTarget(getFieldValue(permission.target));
  const targetType = getFieldValue(permission.target_type);

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Remove Permission?</AlertDialogTitle>
          <AlertDialogDescription className="space-y-2">
            <p>
              Are you sure you want to remove this permission from the role?
            </p>
            <div className="bg-muted rounded-md p-3 text-sm">
              <div className="font-medium">Permission Details:</div>
              <div className="mt-1 space-y-1">
                <div>
                  <span className="text-muted-foreground">Resource Type: </span>
                  {targetType}
                </div>
                <div>
                  <span className="text-muted-foreground">Target: </span>
                  {targetDisplay}
                </div>
                <div>
                  <span className="text-muted-foreground">Kind: </span>
                  {permission.kind}
                </div>
              </div>
            </div>
            <p className="text-muted-foreground text-sm">
              This action cannot be undone.
            </p>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={mutation.isPending}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            onClick={handleDelete}
            disabled={mutation.isPending}
          >
            {mutation.isPending ? (
              <>
                <span>Removing...</span>
              </>
            ) : (
              <>
                <Trash2 className="size-4" />
                Remove Permission
              </>
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
