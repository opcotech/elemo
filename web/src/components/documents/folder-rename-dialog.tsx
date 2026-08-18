import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
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
import { v1FolderUpdate } from "@/lib/api/sdk";
import type { Folder, FolderPatch } from "@/lib/api/types";
import { zFolderCreate } from "@/lib/client/zod.gen";
import { invalidateLibraryQueries } from "@/lib/documents/document-queries";
import { createFormSchema } from "@/lib/forms";

const folderRenameFormSchema = createFormSchema(
  zFolderCreate.pick({ name: true })
);

type FolderRenameFormValues = z.infer<typeof folderRenameFormSchema>;

export function FolderRenameDialog({
  folder,
  open,
  onOpenChange,
}: {
  folder: Pick<Folder, "id" | "name"> | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const form = useForm<FolderRenameFormValues>({
    resolver: zodResolver(folderRenameFormSchema),
    defaultValues: { name: folder?.name ?? "" },
  });
  const queryClient = useQueryClient();

  useEffect(() => {
    if (open) {
      form.reset({ name: folder?.name ?? "" });
    }
  }, [folder, form, open]);

  const mutation = useFormMutation<
    Folder,
    { path: { id: string }; body: FolderPatch },
    FolderRenameFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1FolderUpdate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Folder renamed",
    errorMessagePrefix: "Failed to rename folder",
    transformValues: (values) => ({
      path: { id: folder?.id ?? "" },
      body: { name: values.name.trim() },
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
      title="Rename folder"
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error}
      submitButtonText="Rename"
      onReset={() => form.reset({ name: folder?.name ?? "" })}
      data-section="folder-rename"
    >
      <ControlledField
        control={form.control}
        name="name"
        render={({ field }) => (
          <Field>
            <FieldLabel>Name</FieldLabel>
            <FieldControl>
              <Input autoFocus {...field} />
            </FieldControl>
            <FieldError />
          </Field>
        )}
      />
    </DialogForm>
  );
}
