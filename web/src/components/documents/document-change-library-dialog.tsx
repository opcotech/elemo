import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { DialogForm } from "@/components/ui/dialog-form";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { useAccessibleNamespaces } from "@/lib/api/accessible-namespaces";
import { v1DocumentUpdate } from "@/lib/api/sdk";
import type { Document, DocumentLibrary, DocumentPatch } from "@/lib/api/types";
import { invalidateDocumentQueries } from "@/lib/documents/document-queries";

const changeLibraryFormSchema = z.object({
  library_id: z.string().min(1, "Select a library"),
});

type ChangeLibraryFormValues = z.infer<typeof changeLibraryFormSchema>;

function libraryOptionLabel(library: {
  type: DocumentLibrary["type"];
  name: string;
}): string {
  return `${library.type === "Organization" ? "Organization" : "Namespace"} · ${library.name}`;
}

export function DocumentChangeLibraryDialog({
  documentId,
  currentLibrary,
  open,
  onOpenChange,
}: {
  documentId: string;
  currentLibrary: DocumentLibrary;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const { data: workspace, isLoading } = useAccessibleNamespaces();
  const organizations = workspace?.organizations ?? [];
  const namespaces = workspace?.namespaces ?? [];
  const form = useForm<ChangeLibraryFormValues>({
    resolver: zodResolver(changeLibraryFormSchema),
    defaultValues: {
      library_id: currentLibrary.id,
    },
  });

  useEffect(() => {
    if (open) {
      form.reset({ library_id: currentLibrary.id });
    }
  }, [currentLibrary.id, form, open]);

  const mutation = useFormMutation<
    Document,
    { path: { id: string }; body: DocumentPatch },
    ChangeLibraryFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1DocumentUpdate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Library updated",
    errorMessagePrefix: "Failed to change library",
    transformValues: (values) => ({
      path: { id: documentId },
      body: { library_id: values.library_id },
    }),
    onSuccess: async (document) => {
      await invalidateDocumentQueries(queryClient, document.id);
      onOpenChange(false);
    },
  });

  const items: Record<string, string> = {
    ...Object.fromEntries(
      organizations.map((organization) => [
        organization.id,
        libraryOptionLabel({ type: "Organization", name: organization.name }),
      ])
    ),
    ...Object.fromEntries(
      namespaces.map((namespace) => [
        namespace.id,
        libraryOptionLabel({ type: "Namespace", name: namespace.name }),
      ])
    ),
  };

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Change library"
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error}
      submitButtonText="Move to library"
      onReset={() => form.reset({ library_id: currentLibrary.id })}
      data-section="document-change-library"
    >
      <p className="text-muted-foreground text-sm">
        Moving the document to another library places it at that library root
        and clears its folder.
      </p>
      {isLoading ? (
        <Skeleton className="h-9 w-full" />
      ) : (
        <ControlledField
          control={form.control}
          name="library_id"
          render={({ field }) => (
            <Field>
              <FieldLabel>Library</FieldLabel>
              <Select
                value={field.value}
                onValueChange={field.onChange}
                items={items}
              >
                <FieldControl>
                  <SelectTrigger className="w-full" aria-label="Library">
                    <SelectValue />
                  </SelectTrigger>
                </FieldControl>
                <SelectContent>
                  {organizations.length > 0 ? (
                    <SelectGroup>
                      {organizations.map((organization) => (
                        <SelectItem
                          key={organization.id}
                          value={organization.id}
                        >
                          {libraryOptionLabel({
                            type: "Organization",
                            name: organization.name,
                          })}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  ) : null}
                  {namespaces.length > 0 ? (
                    <SelectGroup>
                      {namespaces.map((namespace) => (
                        <SelectItem key={namespace.id} value={namespace.id}>
                          {libraryOptionLabel({
                            type: "Namespace",
                            name: namespace.name,
                          })}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  ) : null}
                </SelectContent>
              </Select>
              <FieldError />
            </Field>
          )}
        />
      )}
    </DialogForm>
  );
}
