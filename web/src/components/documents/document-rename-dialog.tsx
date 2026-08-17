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
import { v1DocumentUpdate } from "@/lib/api/sdk";
import type { Document, DocumentPatch, PartialDocument } from "@/lib/api/types";
import { zDocumentCreate } from "@/lib/client/zod.gen";
import { invalidateDocumentQueries } from "@/lib/documents/document-queries";
import { createFormSchema } from "@/lib/forms";

const documentRenameFormSchema = createFormSchema(
  zDocumentCreate.pick({ title: true })
);

type DocumentRenameFormValues = z.infer<typeof documentRenameFormSchema>;

export function DocumentRenameDialog({
  document,
  open,
  onOpenChange,
}: {
  document: Pick<PartialDocument, "id" | "title"> | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const form = useForm<DocumentRenameFormValues>({
    resolver: zodResolver(documentRenameFormSchema),
    defaultValues: { title: document?.title ?? "" },
  });
  const queryClient = useQueryClient();

  useEffect(() => {
    if (open) {
      form.reset({ title: document?.title ?? "" });
    }
  }, [document, form, open]);

  const mutation = useFormMutation<
    Document,
    { path: { id: string }; body: DocumentPatch },
    DocumentRenameFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1DocumentUpdate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Document renamed",
    errorMessagePrefix: "Failed to rename document",
    transformValues: (values) => ({
      path: { id: document?.id ?? "" },
      body: { title: values.title.trim() },
    }),
    onSuccess: async (updated) => {
      await invalidateDocumentQueries(queryClient, updated.id);
      onOpenChange(false);
    },
  });

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Rename document"
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error}
      submitButtonText="Rename"
      onReset={() => form.reset({ title: document?.title ?? "" })}
      data-section="document-rename"
    >
      <ControlledField
        control={form.control}
        name="title"
        render={({ field }) => (
          <Field>
            <FieldLabel>Title</FieldLabel>
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
