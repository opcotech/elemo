import { useQueryClient } from "@tanstack/react-query";

import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { useDeleteMutation } from "@/hooks/use-delete-mutation";
import { v1FolderDeleteMutation } from "@/lib/api/mutation-options";
import type { Folder } from "@/lib/api/types";
import { invalidateLibraryQueries } from "@/lib/documents/document-queries";

export function FolderDeleteDialog({
  folder,
  open,
  onOpenChange,
}: {
  folder: Pick<Folder, "id" | "name"> | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const deleteMutation = useDeleteMutation({
    mutationOptions: v1FolderDeleteMutation(),
    successMessage: "Folder deleted",
    successDescription: "Nested folders and documents moved up one level",
    errorMessagePrefix: "Failed to delete folder",
    onSuccess: async () => {
      await invalidateLibraryQueries(queryClient);
      onOpenChange(false);
    },
  });

  return (
    <DeleteConfirmationDialog
      open={open && folder != null}
      onOpenChange={onOpenChange}
      title={
        folder
          ? `Are you sure you want to delete ${folder.name}?`
          : "Delete folder"
      }
      description="This deletes the folder only. Nested folders and documents are not deleted — they move up one level into the parent folder, or to the library root."
      consequences={[
        "The folder will be permanently deleted",
        "Child folders and documents will move up one level",
        "Documents in this folder will not be deleted",
      ]}
      deleteButtonText="Delete"
      onConfirm={() => {
        if (!folder) {
          return;
        }
        deleteMutation.mutate({ path: { id: folder.id } });
      }}
      isPending={deleteMutation.isPending}
    />
  );
}
