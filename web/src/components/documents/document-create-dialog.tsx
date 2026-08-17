import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import type { QueryKey } from "@tanstack/react-query";
import { useForm } from "react-hook-form";

import { DocumentCreateFields } from "@/components/documents/document-create-fields";
import { DialogForm } from "@/components/ui/dialog-form";
import { useFormMutation } from "@/hooks/use-form-mutation";
import type { Document, DocumentCreate } from "@/lib/api/types";
import {
  documentCreateBody,
  documentCreateFormDefaults,
  documentCreateFormSchema,
} from "@/lib/documents/create";
import type { DocumentCreateFormValues } from "@/lib/documents/create";
import { invalidateDocumentQueries } from "@/lib/documents/document-queries";

export function DocumentCreateDialog({
  open,
  onOpenChange,
  create,
  queryKeysToInvalidate,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  create: (body: DocumentCreate) => Promise<Document>;
  queryKeysToInvalidate?: QueryKey[];
}) {
  const form = useForm<DocumentCreateFormValues>({
    resolver: zodResolver(documentCreateFormSchema),
    defaultValues: documentCreateFormDefaults,
  });
  const queryClient = useQueryClient();

  const mutation = useFormMutation<
    Document,
    DocumentCreate,
    DocumentCreateFormValues
  >({
    mutationFn: create,
    form,
    successMessage: "Document created successfully",
    errorMessagePrefix: "Failed to create document",
    resetFormOnSuccess: true,
    queryKeysToInvalidate,
    transformValues: (values) => documentCreateBody(values),
    navigateOnSuccess: (navigate, document) =>
      navigate({
        to: "/documents/$documentId",
        params: { documentId: document.id },
      }),
    onSuccess: async (document) => {
      await invalidateDocumentQueries(queryClient, document.id);
      onOpenChange(false);
    },
  });

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Create document"
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error}
      submitButtonText="Create document"
      onReset={() => form.reset(documentCreateFormDefaults)}
      data-section="document-create"
    >
      <DocumentCreateFields control={form.control} />
    </DialogForm>
  );
}
