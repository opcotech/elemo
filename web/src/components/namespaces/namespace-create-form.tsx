import { zodResolver } from "@hookform/resolvers/zod";
import { useQueries, useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useMemo } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { EntitySelect } from "@/components/ui/entity-select";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldProvider,
} from "@/components/ui/field";
import { FormCard } from "@/components/ui/form-card";
import { Input } from "@/components/ui/input";
import { NameDescriptionFields } from "@/components/ui/name-description-fields";
import { Skeleton } from "@/components/ui/skeleton";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { ResourceType, withResourceType } from "@/hooks/use-permissions";
import { useSuggestedSlug } from "@/hooks/use-suggested-slug";
import { accessibleNamespacesQueryKey } from "@/lib/api/accessible-namespaces";
import { isConflict, throwIfApiFailed } from "@/lib/api/errors";
import {
  v1OrganizationsGetOptions,
  v1OrganizationsNamespacesGetOptions,
  v1PermissionResourceGetOptions,
} from "@/lib/api/query-options";
import { organizationRefPath } from "@/lib/api/refs";
import { zNamespaceCreate } from "@/lib/api/schemas";
import { v1OrganizationsNamespacesCreate } from "@/lib/api/sdk";
import type {
  NamespaceCreate,
  Options,
  Organization,
  V1OrganizationsNamespacesCreateData,
} from "@/lib/api/types";
import { Action, can } from "@/lib/auth/permissions";
import { createFormSchema, normalizeFormData } from "@/lib/forms";
import { namespaceSlugFormSchema } from "@/lib/slug";

const namespaceFormSchema = createFormSchema(
  zNamespaceCreate.omit({ slug: true })
).extend({
  slug: namespaceSlugFormSchema,
});
const namespaceCreateWithOrgSchema = namespaceFormSchema.extend({
  organizationId: z.string().min(1, "Organization is required"),
});

type NamespaceCreateWithOrgFormValues = z.infer<
  typeof namespaceCreateWithOrgSchema
>;

interface NamespaceCreateFormProps {
  organizationId?: string;
  organizationSlug?: string;
}

export function NamespaceCreateForm({
  organizationId,
  organizationSlug,
}: NamespaceCreateFormProps) {
  const navigate = useNavigate();
  const showOrganizationSelector = !organizationId;

  const { data: organizationsPage, isLoading: isLoadingOrgs } = useQuery({
    ...v1OrganizationsGetOptions(),
    enabled: showOrganizationSelector,
  });
  const organizations: Organization[] = organizationsPage?.items ?? [];

  const permissionQueries = useQueries({
    queries: (showOrganizationSelector ? organizations : []).map((org) =>
      v1PermissionResourceGetOptions({
        path: {
          resourceId: withResourceType(ResourceType.Organization, org.id),
        },
      })
    ),
  });

  const writableOrganizations = useMemo(() => {
    if (!showOrganizationSelector) return [];
    return organizations.filter((org, index) => {
      const permissions = permissionQueries[index]?.data;
      return can(permissions, Action.NamespaceCreate);
    });
  }, [showOrganizationSelector, organizations, permissionQueries]);

  const form = useForm<NamespaceCreateWithOrgFormValues>({
    resolver: zodResolver(namespaceCreateWithOrgSchema),
    defaultValues: {
      name: "",
      slug: "",
      description: "",
      organizationId: organizationId ?? "",
    },
  });
  useSuggestedSlug(form);

  const selectedOrganizationId =
    organizationId ?? form.watch("organizationId") ?? undefined;
  const selectedOrganizationSlug =
    organizationSlug ??
    writableOrganizations.find((org) => org.id === selectedOrganizationId)
      ?.slug;

  const mutation = useFormMutation<
    { id: string },
    Options<V1OrganizationsNamespacesCreateData>,
    NamespaceCreateWithOrgFormValues
  >({
    mutationFn: async (variables) => {
      return throwIfApiFailed(
        await v1OrganizationsNamespacesCreate({
          ...variables,
          throwOnError: false,
        })
      );
    },
    form,
    successMessage: "Namespace created",
    errorMessagePrefix: "Failed to create namespace",
    queryKeysToInvalidate: selectedOrganizationId
      ? [
          v1OrganizationsNamespacesGetOptions({
            path: organizationRefPath(selectedOrganizationId),
          }).queryKey,
          v1OrganizationsGetOptions().queryKey,
          accessibleNamespacesQueryKey,
        ]
      : [v1OrganizationsGetOptions().queryKey, accessibleNamespacesQueryKey],
    onError: (error) => {
      if (isConflict(error)) {
        form.setError("slug", {
          type: "conflict",
          message: "This slug is already in use in this organization",
        });
      }
    },
    navigateOnSuccess: (navigateTo) => {
      const slug = form.getValues("slug");
      if (selectedOrganizationSlug) {
        return navigateTo({
          to: "/settings/organizations/$organizationSlug/namespaces/$namespaceSlug",
          params: {
            organizationSlug: selectedOrganizationSlug,
            namespaceSlug: slug,
          },
        });
      }
      return navigateTo({ to: "/settings/namespaces" });
    },
    transformValues: (values) => {
      const { organizationId: formOrganizationId, ...namespaceValues } = values;
      const normalizedBody = normalizeFormData(
        namespaceFormSchema,
        namespaceValues
      ) as NamespaceCreate;
      return {
        path: organizationRefPath(organizationId ?? formOrganizationId),
        body: normalizedBody,
      };
    },
  });

  const handleCancel = () => {
    if (organizationSlug) {
      navigate({
        to: "/settings/organizations/$organizationSlug",
        params: { organizationSlug },
      });
      return;
    }
    navigate({ to: "/settings/namespaces" });
  };

  const isLoadingPermissions = permissionQueries.some((q) => q.isLoading);
  const isLoading =
    showOrganizationSelector && (isLoadingOrgs || isLoadingPermissions);

  return (
    <FormCard
      data-section="namespace-create-form"
      description={
        showOrganizationSelector
          ? "Enter the namespace details below to create a new namespace. Select the organization where the namespace will be created."
          : "Enter the namespace details below to create a new namespace."
      }
      onSubmit={mutation.handleSubmit}
      onCancel={handleCancel}
      isPending={mutation.isPending}
      error={mutation.error || null}
      submitButtonText="Create Namespace"
    >
      <FieldProvider {...form}>
        <FieldGroup>
          {showOrganizationSelector && (
            <ControlledField
              control={form.control}
              name="organizationId"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Organization</FieldLabel>
                  <FieldControl>
                    {isLoading ? (
                      <Skeleton className="h-10 w-full" />
                    ) : writableOrganizations.length > 0 ? (
                      <EntitySelect
                        options={writableOrganizations.map((org) => ({
                          value: org.id,
                          title: org.name,
                          description: org.email || org.website || undefined,
                          avatarSrc: org.logo || null,
                          avatarFallback: org.name,
                        }))}
                        value={field.value}
                        onValueChange={field.onChange}
                        placeholder="Select an organization"
                        disabled={mutation.isPending}
                      />
                    ) : (
                      <div className="text-muted-foreground text-sm">
                        No organizations available
                      </div>
                    )}
                  </FieldControl>
                  <FieldError />
                </Field>
              )}
            />
          )}

          <NameDescriptionFields
            control={form.control}
            isPending={mutation.isPending}
            namePlaceholder="Enter namespace name"
            descriptionPlaceholder="Enter namespace description"
          />

          <ControlledField
            control={form.control}
            name="slug"
            render={({ field }) => (
              <Field>
                <FieldLabel>Slug</FieldLabel>
                <FieldControl>
                  <Input placeholder="platform" {...field} />
                </FieldControl>
                <FieldDescription>
                  Unique within the organization. Cannot be changed later.
                </FieldDescription>
                <FieldError />
              </Field>
            )}
          />
        </FieldGroup>
      </FieldProvider>
    </FormCard>
  );
}
