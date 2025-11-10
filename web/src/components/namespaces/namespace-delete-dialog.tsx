import type { QueryKey } from "@tanstack/react-query";

import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { useDeleteMutation } from "@/hooks/use-delete-mutation";
import type { Namespace } from "@/lib/api";
import {
  v1NamespaceDeleteMutation,
  v1NamespaceGetOptions,
  v1OrganizationsNamespacesGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";

interface NamespaceDeleteDialogProps {
  namespace: Namespace;
  organizationId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function NamespaceDeleteDialog({
  namespace,
  organizationId,
  open,
  onOpenChange,
  onSuccess,
}: NamespaceDeleteDialogProps) {
  const queryKeysToInvalidate: QueryKey[] = [
    v1OrganizationsNamespacesGetOptions({
      path: { id: organizationId },
    }).queryKey,
    v1NamespaceGetOptions({
      path: { id: namespace.id },
    }).queryKey,
  ];

  const deleteMutation = useDeleteMutation({
    mutationOptions: v1NamespaceDeleteMutation(),
    successMessage: "Namespace deleted",
    successDescription: "The namespace has been deleted successfully",
    errorMessagePrefix: "Failed to delete namespace",
    queryKeysToInvalidate,
    onSuccess: () => {
      onSuccess?.();
      onOpenChange(false);
    },
  });

  const handleConfirm = () => {
    deleteMutation.mutate({
      path: {
        id: namespace.id,
      },
    });
  };

  return (
    <DeleteConfirmationDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Are you sure you want to delete ${namespace.name}?`}
      description="This will permanently delete the namespace. This action cannot be undone."
      consequences={[
        "The namespace will be permanently deleted",
        "Projects and documents in this namespace will remain but will no longer be associated with the namespace",
      ]}
      deleteButtonText="Delete"
      onConfirm={handleConfirm}
      isPending={deleteMutation.isPending}
    />
  );
}
