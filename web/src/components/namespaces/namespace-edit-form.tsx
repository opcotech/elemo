import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { FieldGroup, FieldProvider } from "@/components/ui/field";
import { FormCard } from "@/components/ui/form-card";
import { ImmutableIdentifierField } from "@/components/ui/immutable-identifier-field";
import { NameDescriptionFields } from "@/components/ui/name-description-fields";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { accessibleNamespacesQueryKey } from "@/lib/api/accessible-namespaces";
import {
  v1NamespaceGetOptions,
  v1OrganizationsNamespacesGetOptions,
} from "@/lib/api/query-options";
import { namespaceRefPath, organizationRefPath } from "@/lib/api/refs";
import { zNamespacePatch } from "@/lib/api/schemas";
import { v1NamespaceUpdate } from "@/lib/api/sdk";
import type {
  Namespace,
  Options,
  V1NamespaceUpdateData,
} from "@/lib/api/types";
import { createFormSchema, normalizePatchData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

const namespaceEditFormSchema = createFormSchema(zNamespacePatch);

type NamespaceEditFormValues = z.infer<typeof namespaceEditFormSchema>;

interface NamespaceEditFormProps {
  namespace: Namespace & { slug: string };
  organizationId: string;
  organizationSlug: string;
}

export function NamespaceEditForm({
  namespace,
  organizationId,
  organizationSlug,
}: NamespaceEditFormProps) {
  const navigate = useNavigate();
  const organizationRoute = {
    to: "/settings/organizations/$organizationSlug",
    params: { organizationSlug },
  } as const;

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
        path: organizationRefPath(organizationId),
      }).queryKey,
      v1OrganizationsNamespacesGetOptions({
        path: organizationRefPath(organizationSlug),
      }).queryKey,
      v1NamespaceGetOptions({
        path: namespaceRefPath(organizationId, namespace.id),
      }).queryKey,
      v1NamespaceGetOptions({
        path: namespaceRefPath(organizationSlug, namespace.slug),
      }).queryKey,
      accessibleNamespacesQueryKey,
    ],
    navigateOnSuccess: (navigateTo) => navigateTo(organizationRoute),
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
        path: namespaceRefPath(organizationId, namespace.id),
        body: normalizedBody,
      };
    },
  });

  return (
    <FormCard
      description="Update the namespace details below."
      onSubmit={mutation.handleSubmit}
      onCancel={() => navigate(organizationRoute)}
      isPending={mutation.isPending}
      error={mutation.error || null}
    >
      <FieldProvider {...form}>
        <FieldGroup>
          <ImmutableIdentifierField label="Slug" value={namespace.slug} />
          <NameDescriptionFields
            control={form.control}
            isPending={mutation.isPending}
            namePlaceholder="Enter namespace name"
            descriptionPlaceholder="Enter namespace description"
          />
        </FieldGroup>
      </FieldProvider>
    </FormCard>
  );
}
