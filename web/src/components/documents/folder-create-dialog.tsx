import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { DialogForm } from "@/components/ui/dialog-form";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { zFolderCreate } from "@/lib/api/schemas";
import {
  v1NamespacesFoldersCreate,
  v1OrganizationsFoldersCreate,
} from "@/lib/api/sdk";
import type { Folder, FolderCreate } from "@/lib/api/types";
import { invalidateLibraryQueries } from "@/lib/documents/document-queries";
import type { DocumentLibraryKind } from "@/lib/documents/library";
import { createFormSchema } from "@/lib/forms";

const folderCreateFormSchema = createFormSchema(
  zFolderCreate.pick({ name: true })
);

type FolderCreateFormValues = z.infer<typeof folderCreateFormSchema>;

const folderCreateFormDefaults: FolderCreateFormValues = {
  name: "",
};

export function FolderCreateDialog({
  kind,
  libraryId,
  parentId,
  open,
  onOpenChange,
}: {
  kind: DocumentLibraryKind;
  libraryId: string;
  parentId?: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const form = useForm<FolderCreateFormValues>({
    resolver: zodResolver(folderCreateFormSchema),
    defaultValues: folderCreateFormDefaults,
  });
  const queryClient = useQueryClient();

  const mutation = useFormMutation<
    Folder,
    { path: { id: string }; body: FolderCreate },
    FolderCreateFormValues
  >({
    mutationFn: async (variables) => {
      const create =
        kind === "organization"
          ? v1OrganizationsFoldersCreate
          : v1NamespacesFoldersCreate;
      const { data } = await create({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Folder created",
    errorMessagePrefix: "Failed to create folder",
    resetFormOnSuccess: true,
    transformValues: (values) => ({
      path: { id: libraryId },
      body: {
        name: values.name.trim(),
        ...(parentId ? { parent_id: parentId } : {}),
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
      title="New folder"
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error}
      submitButtonText="Create folder"
      onReset={() => form.reset(folderCreateFormDefaults)}
      data-section="folder-create"
    >
      <ControlledField
        control={form.control}
        name="name"
        render={({ field }) => (
          <Field>
            <FieldLabel>Name</FieldLabel>
            <FieldControl>
              <Input autoFocus placeholder="Guides" {...field} />
            </FieldControl>
            <FieldError />
          </Field>
        )}
      />
    </DialogForm>
  );
}
