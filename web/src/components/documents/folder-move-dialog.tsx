import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { FolderPickerField } from "./folder-picker";

import { DialogForm } from "@/components/ui/dialog-form";
import { Skeleton } from "@/components/ui/skeleton";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { v1FolderUpdate } from "@/lib/api/sdk";
import type { Folder, FolderPatch } from "@/lib/api/types";
import { invalidateLibraryQueries } from "@/lib/documents/document-queries";
import {
  LIBRARY_ROOT_FOLDER_VALUE,
  folderMoveTargets,
  libraryFolderOptionsQuery,
  libraryFolderPickerOptions,
} from "@/lib/documents/library";
import type { DocumentLibraryKind } from "@/lib/documents/library";

const moveFormSchema = z.object({
  parent_id: z.string().min(1),
});

type MoveFormValues = z.infer<typeof moveFormSchema>;

export function FolderMoveDialog({
  folder,
  kind,
  organizationId,
  namespaceId,
  open,
  onOpenChange,
}: {
  folder: Pick<Folder, "id" | "name" | "parent"> | null;
  kind: DocumentLibraryKind;
  organizationId: string;
  namespaceId?: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const { data: folders, isLoading } = useQuery({
    ...libraryFolderOptionsQuery(kind, organizationId, namespaceId),
    enabled: open && folder != null,
  });
  const currentParentId = folder?.parent?.id ?? LIBRARY_ROOT_FOLDER_VALUE;
  const form = useForm<MoveFormValues>({
    resolver: zodResolver(moveFormSchema),
    defaultValues: {
      parent_id: currentParentId,
    },
  });
  const targets = useMemo(
    () => (folder ? folderMoveTargets(folders ?? [], folder.id) : []),
    [folder, folders]
  );
  const options = libraryFolderPickerOptions(targets);

  useEffect(() => {
    if (open) {
      form.reset({
        parent_id: currentParentId,
      });
    }
  }, [currentParentId, form, open]);

  const mutation = useFormMutation<
    Folder,
    { path: { id: string }; body: FolderPatch },
    MoveFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1FolderUpdate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Folder moved",
    errorMessagePrefix: "Failed to move folder",
    transformValues: (values) => ({
      path: { id: folder?.id ?? "" },
      body: {
        parent_id:
          values.parent_id === LIBRARY_ROOT_FOLDER_VALUE
            ? null
            : values.parent_id,
      },
    }),
    onSuccess: async () => {
      await invalidateLibraryQueries(queryClient);
      onOpenChange(false);
    },
  });

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Move folder"
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error}
      submitButtonText="Move"
      onReset={() =>
        form.reset({
          parent_id: currentParentId,
        })
      }
      data-section="folder-move"
    >
      <p className="text-muted-foreground text-sm">
        Choose a folder in this library for{" "}
        <span className="text-foreground font-medium">
          {folder?.name ?? "this folder"}
        </span>
        . A folder cannot be moved into itself or one of its nested folders.
      </p>
      {isLoading ? (
        <Skeleton className="h-9 w-full" />
      ) : (
        <FolderPickerField
          control={form.control}
          name="parent_id"
          options={options}
          disabled={mutation.isPending}
        />
      )}
    </DialogForm>
  );
}
