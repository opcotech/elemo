import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { PlusIcon } from "lucide-react";
import { useForm } from "react-hook-form";

import { DocumentCreateFields } from "@/components/documents/document-create-fields";
import { QuickCreateContext } from "@/components/quick-create/context-panel";
import type { QuickCreateKindProps } from "@/components/quick-create/types";
import { MockDataAlert } from "@/components/shared/app-feedback";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { DialogFooter } from "@/components/ui/dialog";
import { FieldGroup, FieldProvider } from "@/components/ui/field";
import { Spinner } from "@/components/ui/spinner";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { useNavigationContext } from "@/hooks/use-navigation-context";
import {
  v1NamespacesDocumentsCreate,
  v1OrganizationsDocumentsCreate,
  v1ProjectsDocumentsCreate,
} from "@/lib/api/sdk";
import type { Document, DocumentCreate } from "@/lib/api/types";
import {
  documentCreateBody,
  documentCreateFormDefaults,
  documentCreateFormSchema,
  documentCreateParentFromNavigation,
  documentListQueryKey,
} from "@/lib/documents/create";
import type { DocumentCreateFormValues } from "@/lib/documents/create";
import { invalidateDocumentQueries } from "@/lib/documents/document-queries";

export function DocumentQuickCreate({
  onCancel,
  onComplete,
}: QuickCreateKindProps) {
  const navigation = useNavigationContext();
  const queryClient = useQueryClient();
  const parent = documentCreateParentFromNavigation(navigation);
  const form = useForm<DocumentCreateFormValues>({
    resolver: zodResolver(documentCreateFormSchema),
    defaultValues: documentCreateFormDefaults,
  });

  const mutation = useFormMutation<
    Document,
    { path: { id: string }; body: DocumentCreate },
    DocumentCreateFormValues
  >({
    mutationFn: async (variables) => {
      const create =
        parent?.type === "project"
          ? v1ProjectsDocumentsCreate
          : parent?.type === "namespace"
            ? v1NamespacesDocumentsCreate
            : v1OrganizationsDocumentsCreate;
      const { data } = await create({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Document created successfully",
    errorMessagePrefix: "Failed to create document",
    resetFormOnSuccess: true,
    queryKeysToInvalidate: parent
      ? [documentListQueryKey(parent.type, parent.id)]
      : [],
    transformValues: (values) => {
      return {
        path: { id: parent!.id },
        body: documentCreateBody(values),
      };
    },
    navigateOnSuccess: (navigate, document) =>
      navigate({
        to: "/documents/$documentId",
        params: { documentId: document.id },
      }),
    onSuccess: async (document) => {
      await invalidateDocumentQueries(queryClient, document.id);
      onComplete();
    },
  });

  const errorMessage = mutation.error?.message;

  if (!parent) {
    return (
      <FieldProvider {...form}>
        <form
          onSubmit={(event) => {
            event.preventDefault();
          }}
        >
          <FieldGroup className="my-5">
            <DocumentCreateFields control={form.control} />
            <QuickCreateContext />
            <MockDataAlert title="Parent context required">
              Documents can be created from an organization, namespace, or
              project. Open one first, then use quick create.
            </MockDataAlert>
          </FieldGroup>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button type="submit" disabled>
              <PlusIcon />
              Create unavailable
            </Button>
          </DialogFooter>
        </form>
      </FieldProvider>
    );
  }

  return (
    <FieldProvider {...form}>
      <form onSubmit={mutation.handleSubmit}>
        <FieldGroup className="my-5">
          {errorMessage && (
            <Alert variant="destructive">
              <AlertTitle>Failed to save</AlertTitle>
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>
          )}

          <DocumentCreateFields control={form.control} />
          <QuickCreateContext />
        </FieldGroup>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={mutation.isPending}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? (
              <>
                <Spinner size="xs" className="mr-0.5 text-white" />
                <span>Saving...</span>
              </>
            ) : (
              <>
                <PlusIcon />
                Create document
              </>
            )}
          </Button>
        </DialogFooter>
      </form>
    </FieldProvider>
  );
}
