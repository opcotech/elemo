import type { QueryKey } from "@tanstack/react-query";

import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { useDeleteMutation } from "@/hooks/use-delete-mutation";
import type { Organization } from "@/lib/api";
import {
  v1OrganizationDeleteMutation,
  v1OrganizationGetOptions,
  v1OrganizationsGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";

interface OrganizationDeleteDialogProps {
  organization: Organization;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function OrganizationDeleteDialog({
  organization,
  open,
  onOpenChange,
  onSuccess,
}: OrganizationDeleteDialogProps) {
  const queryKeysToInvalidate: QueryKey[] = [
    v1OrganizationGetOptions({
      path: { id: organization.id },
    }).queryKey,
    v1OrganizationsGetOptions().queryKey,
  ];

  const deleteMutation = useDeleteMutation({
    mutationOptions: v1OrganizationDeleteMutation(),
    successMessage: "Organization deleted",
    successDescription: "The organization has been deleted successfully",
    errorMessagePrefix: "Failed to delete organization",
    queryKeysToInvalidate,
    onSuccess: () => {
      onSuccess?.();
      onOpenChange(false);
    },
    navigateOnSuccess: "/settings/organizations",
  });

  const handleConfirm = () => {
    deleteMutation.mutate({
      path: { id: organization.id },
      query: { force: false }, // Soft delete (default)
    });
  };

  return (
    <DeleteConfirmationDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Are you sure you want to delete ${organization.name}?`}
      description="This will mark the organization as deleted. This action cannot be undone."
      consequences={[
        "The organization will be marked as deleted",
        "All organization members will lose access",
        "Organization data will be hidden from listings",
        "You will be redirected to the organizations list",
      ]}
      deleteButtonText="Delete"
      onConfirm={handleConfirm}
      isPending={deleteMutation.isPending}
    />
  );
}
