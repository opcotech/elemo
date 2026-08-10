import type { QueryKey } from "@tanstack/react-query";

import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { useDeleteMutation } from "@/hooks/use-delete-mutation";
import type { PartialProject, Project } from "@/lib/api";
import {
  v1NamespaceGetOptions,
  v1NamespacesProjectsGetOptions,
  v1ProjectDeleteMutation,
  v1ProjectGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";

interface ProjectDeleteDialogProps {
  project: Pick<Project | PartialProject, "id" | "name">;
  organizationId: string;
  namespaceId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
  navigateOnSuccess?: boolean;
}

export function ProjectDeleteDialog({
  project,
  organizationId,
  namespaceId,
  open,
  onOpenChange,
  onSuccess,
  navigateOnSuccess = false,
}: ProjectDeleteDialogProps) {
  const queryKeysToInvalidate: QueryKey[] = [
    v1ProjectGetOptions({
      path: { id: project.id },
    }).queryKey,
    v1NamespaceGetOptions({
      path: { id: namespaceId },
    }).queryKey,
    v1NamespacesProjectsGetOptions({
      path: { id: namespaceId },
    }).queryKey,
  ];

  const deleteMutation = useDeleteMutation({
    mutationOptions: v1ProjectDeleteMutation(),
    successMessage: "Project deleted",
    successDescription: "The project has been deleted successfully",
    errorMessagePrefix: "Failed to delete project",
    queryKeysToInvalidate,
    onSuccess: () => {
      onSuccess?.();
      onOpenChange(false);
    },
    navigateOnSuccess: navigateOnSuccess
      ? {
          to: "/settings/organizations/$organizationId/namespaces/$namespaceId",
          params: { organizationId, namespaceId },
        }
      : undefined,
  });

  const handleConfirm = () => {
    deleteMutation.mutate({
      path: {
        id: project.id,
      },
    });
  };

  return (
    <DeleteConfirmationDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Are you sure you want to delete ${project.name}?`}
      description="This will permanently delete the project. This action cannot be undone."
      consequences={[
        "The project will be permanently deleted",
        "Documents and issues in this project will remain but will no longer be associated with the project",
      ]}
      deleteButtonText="Delete"
      onConfirm={handleConfirm}
      isPending={deleteMutation.isPending}
    />
  );
}
