import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
} from "@/components/ui/card";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldProvider,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { accessibleNamespacesQueryKey } from "@/lib/api/accessible-namespaces";
import {
  v1NamespaceGetOptions,
  v1OrganizationsNamespacesGetOptions,
} from "@/lib/api/query-options";
import { v1NamespaceUpdate } from "@/lib/api/sdk";
import type {
  Namespace,
  Options,
  V1NamespaceUpdateData,
} from "@/lib/api/types";
import { zNamespacePatch } from "@/lib/client/zod.gen";
import { createFormSchema, normalizePatchData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

const namespaceEditFormSchema = createFormSchema(zNamespacePatch);

type NamespaceEditFormValues = z.infer<typeof namespaceEditFormSchema>;

interface NamespaceEditFormProps {
  namespace: Namespace;
  organizationId: string;
}

export function NamespaceEditForm({
  namespace,
  organizationId,
}: NamespaceEditFormProps) {
  const navigate = useNavigate();
  const form = useForm<NamespaceEditFormValues>({
    resolver: zodResolver(namespaceEditFormSchema),
    defaultValues: {
      name: namespace.name,
      description: getDefaultValue(namespace.description),
    },
  });

  useEffect(() => {
    if (!form.formState.isDirty) {
      form.reset({
        name: namespace.name,
        description: getDefaultValue(namespace.description),
      });
    }
  }, [namespace.name, namespace.description, form]);

  const mutation = useFormMutation<
    Namespace,
    Options<V1NamespaceUpdateData>,
    NamespaceEditFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1NamespaceUpdate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Namespace updated",
    errorMessagePrefix: "Failed to update namespace",
    queryKeysToInvalidate: [
      v1OrganizationsNamespacesGetOptions({
        path: { id: organizationId },
      }).queryKey,
      v1NamespaceGetOptions({
        path: { id: namespace.id },
      }).queryKey,
      accessibleNamespacesQueryKey,
    ],
    navigateOnSuccess: (navigateTo) =>
      navigateTo({
        to: "/settings/organizations/$organizationId",
        params: { organizationId },
      }),
    transformValues: (values) => {
      const normalizedBody = normalizePatchData(
        namespaceEditFormSchema,
        values,
        {
          name: namespace.name,
          description: namespace.description,
        }
      );
      return {
        path: {
          id: namespace.id,
        },
        body: normalizedBody,
      };
    },
  });

  return (
    <Card>
      <CardHeader>
        <CardDescription>Update the namespace details below.</CardDescription>
      </CardHeader>
      <CardContent>
        <FieldProvider {...form}>
          <form
            onSubmit={mutation.handleSubmit}
            className="flex flex-col gap-y-6"
          >
            {mutation.isError && (
              <Alert variant="destructive">
                <AlertTitle>Failed to update namespace</AlertTitle>
                <AlertDescription>{mutation.error?.message}</AlertDescription>
              </Alert>
            )}

            <FieldGroup>
              <ControlledField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <Field>
                    <FieldLabel>Name</FieldLabel>
                    <FieldControl>
                      <Input
                        placeholder="Enter namespace name"
                        {...field}
                        disabled={mutation.isPending}
                      />
                    </FieldControl>
                    <FieldError />
                  </Field>
                )}
              />

              <ControlledField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <Field>
                    <FieldLabel>Description</FieldLabel>
                    <FieldControl>
                      <Textarea
                        placeholder="Enter namespace description (optional)"
                        {...field}
                        value={getDefaultValue(field.value)}
                        rows={4}
                        disabled={mutation.isPending}
                      />
                    </FieldControl>
                    <FieldError />
                  </Field>
                )}
              />
            </FieldGroup>

            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() =>
                  navigate({
                    to: "/settings/organizations/$organizationId",
                    params: { organizationId },
                  })
                }
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
                  "Save Changes"
                )}
              </Button>
            </div>
          </form>
        </FieldProvider>
      </CardContent>
    </Card>
  );
}
