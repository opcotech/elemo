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
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldProvider,
} from "@/components/ui/field";
import { FormCard } from "@/components/ui/form-card";
import { NameDescriptionFields } from "@/components/ui/name-description-fields";
import { Skeleton } from "@/components/ui/skeleton";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { ResourceType, withResourceType } from "@/hooks/use-permissions";
import { accessibleNamespacesQueryKey } from "@/lib/api/accessible-namespaces";
import {
  v1OrganizationsGetOptions,
  v1OrganizationsNamespacesGetOptions,
  v1PermissionResourceGetOptions,
} from "@/lib/api/query-options";
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

const namespaceFormSchema = createFormSchema(zNamespaceCreate);
const namespaceCreateWithOrgSchema = namespaceFormSchema.extend({
  organizationId: z.string().min(1, "Organization is required"),
});

type NamespaceCreateWithOrgFormValues = z.infer<
  typeof namespaceCreateWithOrgSchema
>;

interface NamespaceCreateFormProps {
  organizationId?: string;
}

export function NamespaceCreateForm({
  organizationId,
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
      description: "",
      organizationId: organizationId ?? "",
    },
  });

  const selectedOrganizationId =
    organizationId ?? form.watch("organizationId") ?? undefined;

  const mutation = useFormMutation<
    { id: string },
    Options<V1OrganizationsNamespacesCreateData>,
    NamespaceCreateWithOrgFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1OrganizationsNamespacesCreate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Namespace created",
    errorMessagePrefix: "Failed to create namespace",
    queryKeysToInvalidate: selectedOrganizationId
      ? [
          v1OrganizationsNamespacesGetOptions({
            path: { id: selectedOrganizationId },
          }).queryKey,
          v1OrganizationsGetOptions().queryKey,
          accessibleNamespacesQueryKey,
        ]
      : [v1OrganizationsGetOptions().queryKey, accessibleNamespacesQueryKey],
    navigateOnSuccess: (navigateTo) =>
      organizationId
        ? navigateTo({
            to: "/settings/organizations/$organizationId",
            params: { organizationId },
          })
        : navigateTo({
            to: "/settings/namespaces",
          }),
    transformValues: (values) => {
      const { organizationId: formOrganizationId, ...namespaceValues } = values;
      const normalizedBody = normalizeFormData(
        namespaceFormSchema,
        namespaceValues
      ) as NamespaceCreate;
      return {
        path: {
          id: organizationId ?? formOrganizationId,
        },
        body: normalizedBody,
      };
    },
  });

  const handleCancel = () => {
    if (organizationId) {
      navigate({
        to: "/settings/organizations/$organizationId",
        params: { organizationId },
      });
    } else {
      navigate({
        to: "/settings/namespaces",
      });
    }
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
        </FieldGroup>
      </FieldProvider>
    </FormCard>
  );
}
