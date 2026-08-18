import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { DialogForm } from "@/components/ui/dialog-form";
import { SearchableEntitySelect } from "@/components/ui/entity-select";
import type { EntitySelectOption } from "@/components/ui/entity-select";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Skeleton } from "@/components/ui/skeleton";
import type { PartialDocument } from "@/lib/api/types";
import { invalidateDocumentQueries } from "@/lib/documents/document-queries";
import {
  filterAvailableDocuments,
  relatedDocumentCatalogQueryOptions,
} from "@/lib/documents/link";
import { runMutationSuccessWorkflow } from "@/lib/mutation-workflow";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

const linkFormSchema = z.object({
  documentId: z.string().min(1, "Document is required"),
});

type LinkFormValues = z.infer<typeof linkFormSchema>;

function documentToSelectOption(
  document: Pick<PartialDocument, "id" | "title" | "excerpt">
): EntitySelectOption {
  return {
    value: document.id,
    title: document.title,
    description: document.excerpt || undefined,
    searchText: [document.title, document.excerpt].filter(Boolean).join(" "),
  };
}

export function DocumentLinkDialog({
  namespaceId,
  relatedIds,
  relatedLabel,
  open,
  onOpenChange,
  onLink,
}: {
  namespaceId: string;
  relatedIds: ReadonlySet<string>;
  relatedLabel: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onLink: (documentId: string) => Promise<void>;
}) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const [error, setError] = useState<Error | null>(null);

  const form = useForm<LinkFormValues>({
    resolver: zodResolver(linkFormSchema),
    defaultValues: {
      documentId: "",
    },
  });

  const { data: documentsPage, isLoading } = useQuery({
    ...relatedDocumentCatalogQueryOptions(namespaceId),
    enabled: open && Boolean(namespaceId),
  });

  const availableDocuments = filterAvailableDocuments(
    documentsPage?.items ?? [],
    relatedIds
  );
  const documentOptions = availableDocuments.map(documentToSelectOption);

  const mutation = useMutation({
    mutationFn: async (document: PartialDocument) => {
      await onLink(document.id);
      return document;
    },
    onSuccess: (document) =>
      runMutationSuccessWorkflow({
        invalidateQueries: [
          () => invalidateDocumentQueries(queryClient, document.id),
        ],
        invalidateRouter: () => router.invalidate(),
        callbacks: [
          () => {
            setError(null);
            showSuccessToast(
              "Document linked",
              `${document.title} is now related to ${relatedLabel}`
            );
            form.reset();
            onOpenChange(false);
          },
        ],
      }),
    onError: (err) => {
      const nextError =
        err instanceof Error ? err : new Error("Unknown error occurred");
      setError(nextError);
      showErrorToast("Failed to link document", nextError.message);
    },
  });

  useEffect(() => {
    if (open) {
      setError(null);
    }
  }, [open]);

  const onSubmit = (values: LinkFormValues) => {
    const document = availableDocuments.find(
      (item) => item.id === values.documentId
    );
    if (!document) {
      return;
    }
    setError(null);
    mutation.mutate(document);
  };

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Link document"
      data-section="document-link-form"
      onSubmit={form.handleSubmit(onSubmit)}
      isPending={mutation.isPending}
      error={error}
      submitButtonText="Link document"
      onReset={() => form.reset()}
      className="sm:max-w-125"
    >
      {isLoading ? (
        <div className="space-y-2">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-10 w-full" />
        </div>
      ) : availableDocuments.length === 0 ? (
        <Alert>
          <AlertDescription>
            {relatedIds.size > 0
              ? "All documents in this namespace are already linked."
              : "No documents in this namespace are available to link."}
          </AlertDescription>
        </Alert>
      ) : (
        <ControlledField
          control={form.control}
          name="documentId"
          render={({ field }) => (
            <Field>
              <FieldLabel>Document</FieldLabel>
              <FieldControl>
                <SearchableEntitySelect
                  options={documentOptions}
                  value={field.value}
                  onValueChange={field.onChange}
                  placeholder="Choose a document"
                  searchPlaceholder="Search documents…"
                  emptyMessage="No documents found."
                  disabled={mutation.isPending}
                />
              </FieldControl>
              <FieldError />
            </Field>
          )}
        />
      )}
    </DialogForm>
  );
}
