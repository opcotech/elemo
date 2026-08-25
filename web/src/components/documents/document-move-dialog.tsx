import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { FolderPickerField } from "./folder-picker";

import { DialogForm } from "@/components/ui/dialog-form";
import { Skeleton } from "@/components/ui/skeleton";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { v1DocumentUpdate } from "@/lib/api/sdk";
import type { Document, DocumentPatch } from "@/lib/api/types";
import { invalidateDocumentQueries } from "@/lib/documents/document-queries";
import {
  LIBRARY_ROOT_FOLDER_VALUE,
  libraryFolderOptionsQuery,
  libraryFolderPickerOptions,
} from "@/lib/documents/library";
import type { DocumentLibraryKind } from "@/lib/documents/library";

const moveFormSchema = z.object({
  folder_id: z.string().min(1),
});

type MoveFormValues = z.infer<typeof moveFormSchema>;

export function DocumentMoveDialog({
  documentId,
  documentTitle,
  kind,
  organizationId,
  namespaceId,
  currentFolderId,
  open,
  onOpenChange,
}: {
  documentId: string;
  documentTitle: string;
  kind: DocumentLibraryKind;
  organizationId: string;
  namespaceId?: string;
  currentFolderId?: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const { data: folders, isLoading } = useQuery({
    ...libraryFolderOptionsQuery(kind, organizationId, namespaceId),
    enabled: open,
  });
  const form = useForm<MoveFormValues>({
    resolver: zodResolver(moveFormSchema),
    defaultValues: {
      folder_id: currentFolderId ?? LIBRARY_ROOT_FOLDER_VALUE,
    },
  });

  useEffect(() => {
    if (open) {
      form.reset({
        folder_id: currentFolderId ?? LIBRARY_ROOT_FOLDER_VALUE,
      });
    }
  }, [currentFolderId, form, open]);

  const mutation = useFormMutation<
    Document,
    { path: { id: string }; body: DocumentPatch },
    MoveFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1DocumentUpdate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Document moved",
    errorMessagePrefix: "Failed to move document",
    transformValues: (values) => ({
      path: { id: documentId },
      body: {
        folder_id:
          values.folder_id === LIBRARY_ROOT_FOLDER_VALUE
            ? null
            : values.folder_id,
      },
    }),
    onSuccess: async (document) => {
      await invalidateDocumentQueries(queryClient, document.id);
      onOpenChange(false);
    },
  });

  const options = libraryFolderPickerOptions(folders ?? []);

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Move document"
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error}
      submitButtonText="Move"
      onReset={() =>
        form.reset({
          folder_id: currentFolderId ?? LIBRARY_ROOT_FOLDER_VALUE,
        })
      }
      data-section="document-move"
    >
      <p className="text-muted-foreground text-sm">
        Choose a folder in this library for{" "}
        <span className="text-foreground font-medium">{documentTitle}</span>.
      </p>
      {isLoading ? (
        <Skeleton className="h-9 w-full" />
      ) : (
        <FolderPickerField
          control={form.control}
          name="folder_id"
          options={options}
          disabled={mutation.isPending}
        />
      )}
    </DialogForm>
  );
}
