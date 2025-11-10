import type { QueryKey } from "@tanstack/react-query";

import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { useDeleteMutation } from "@/hooks/use-delete-mutation";
import type { Role } from "@/lib/api";
import {
  v1OrganizationRoleDeleteMutation,
  v1OrganizationRoleGetOptions,
  v1OrganizationRolesGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";

interface OrganizationRoleDeleteDialogProps {
  role: Role;
  organizationId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function OrganizationRoleDeleteDialog({
  role,
  organizationId,
  open,
  onOpenChange,
  onSuccess,
}: OrganizationRoleDeleteDialogProps) {
  const queryKeysToInvalidate: QueryKey[] = [
    v1OrganizationRolesGetOptions({
      path: { id: organizationId },
    }).queryKey,
    v1OrganizationRoleGetOptions({
      path: {
        id: organizationId,
        role_id: role.id,
      },
    }).queryKey,
  ];

  const deleteMutation = useDeleteMutation({
    mutationOptions: v1OrganizationRoleDeleteMutation(),
    successMessage: "Role deleted",
    successDescription: "The role has been deleted successfully",
    errorMessagePrefix: "Failed to delete role",
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
        role_id: role.id,
      },
    });
  };

  return (
    <DeleteConfirmationDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Are you sure you want to delete ${role.name}?`}
      description="This will permanently delete the role. This action cannot be undone."
      consequences={[
        "The role will be permanently deleted",
        "All members assigned to this role will lose their role assignment",
        "Role permissions will be removed",
      ]}
      deleteButtonText="Delete"
      onConfirm={handleConfirm}
      isPending={deleteMutation.isPending}
    />
  );
}
