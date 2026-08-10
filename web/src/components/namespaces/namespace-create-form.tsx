import { zodResolver } from "@hookform/resolvers/zod";
import { useQueries, useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useMemo } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
} from "@/components/ui/card";
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
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { ResourceType, withResourceType } from "@/hooks/use-permissions";
import { accessibleNamespacesQueryKey } from "@/lib/api/accessible-namespaces";
import {
  v1OrganizationsGetOptions,
  v1OrganizationsNamespacesGetOptions,
  v1PermissionResourceGetOptions,
} from "@/lib/api/query-options";
import { v1OrganizationsNamespacesCreate } from "@/lib/api/sdk";
import type {
  NamespaceCreate,
  Options,
  V1OrganizationsNamespacesCreateData,
} from "@/lib/api/types";
import { can } from "@/lib/auth/permissions";
import { zNamespaceCreate } from "@/lib/client/zod.gen";
import { createFormSchema, normalizeFormData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

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

  const { data: organizations, isLoading: isLoadingOrgs } = useQuery({
    ...v1OrganizationsGetOptions(),
    enabled: showOrganizationSelector,
  });

  // Check permissions for each organization (only if showing selector)
  const permissionQueries = useQueries({
    queries:
      showOrganizationSelector && organizations && organizations.length > 0
        ? organizations.map((org) =>
            v1PermissionResourceGetOptions({
              path: {
                resourceId: withResourceType(ResourceType.Organization, org.id),
              },
            })
          )
        : [],
  });

  // Filter organizations to only those where user has write permission
  const writableOrganizations = useMemo(() => {
    if (!showOrganizationSelector || !organizations) return [];
    return organizations.filter((org, index) => {
      const permissions = permissionQueries[index]?.data;
      return can(permissions, "write");
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
    <Card>
      <CardHeader>
        <CardDescription>
          {showOrganizationSelector
            ? "Enter the namespace details below to create a new namespace. Select the organization where the namespace will be created."
            : "Enter the namespace details below to create a new namespace."}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <FieldProvider {...form}>
          <form
            onSubmit={mutation.handleSubmit}
            className="flex flex-col gap-y-6"
          >
            {mutation.isError && (
              <div className="text-destructive text-sm">
                {mutation.error?.message}
              </div>
            )}

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
                        ) : writableOrganizations &&
                          writableOrganizations.length > 0 ? (
                          <EntitySelect
                            options={writableOrganizations.map((org) => ({
                              value: org.id,
                              title: org.name,
                              description:
                                org.email || org.website || undefined,
                              avatarSrc:
                                (org as { logo?: string | null }).logo || null,
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
                onClick={handleCancel}
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
                  "Create Namespace"
                )}
              </Button>
            </div>
          </form>
        </FieldProvider>
      </CardContent>
    </Card>
  );
}
