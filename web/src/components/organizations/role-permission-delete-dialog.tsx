import type { QueryKey } from "@tanstack/react-query";

import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { useDeleteMutation } from "@/hooks/use-delete-mutation";
import type { Permission } from "@/lib/api";
import {
  v1OrganizationRolePermissionRemoveMutation,
  v1OrganizationRolePermissionsGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { extractResourceId, formatResourceId } from "@/lib/utils";

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
  const queryKeysToInvalidate: QueryKey[] = [
    v1OrganizationRolePermissionsGetOptions({
      path: {
        id: organizationId,
        role_id: roleId,
      },
    }).queryKey,
  ];

  const deleteMutation = useDeleteMutation({
    mutationOptions: v1OrganizationRolePermissionRemoveMutation(),
    successMessage: "Permission removed",
    successDescription: "Permission removed successfully",
    errorMessagePrefix: "Failed to remove permission",
    queryKeysToInvalidate,
    onSuccess: () => {
      onSuccess?.();
      onOpenChange(false);
    },
  });

  const handleConfirm = () => {
    deleteMutation.mutate({
      path: {
        id: organizationId,
        role_id: roleId,
        permission_id: permission.id,
      },
    });
  };

  const targetDisplay = formatResourceId(extractResourceId(permission.target));
  const targetType = permission.target_type;

  return (
    <DeleteConfirmationDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Remove Permission?"
      description="Are you sure you want to remove this permission from the role?"
      consequences={[
        "The permission will be removed from this role",
        "This action cannot be undone",
      ]}
      deleteButtonText="Remove Permission"
      onConfirm={handleConfirm}
      isPending={deleteMutation.isPending}
    >
      <div className="bg-primary/5 ring-primary/10 mt-2 rounded-md p-3 text-sm ring-1">
        <div className="space-y-1">
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
    </DeleteConfirmationDialog>
  );
}
