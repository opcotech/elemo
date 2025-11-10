import { zodResolver } from "@hookform/resolvers/zod";
import { useQueries, useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useMemo } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { namespaceFormSchema } from "./namespace-form-fields";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
} from "@/components/ui/card";
import { EntitySelect } from "@/components/ui/entity-select";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { ResourceType, withResourceType } from "@/hooks/use-permissions";
import type {
  NamespaceCreate,
  Options,
  V1OrganizationsNamespacesCreateData,
} from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import {
  v1OrganizationsGetOptions,
  v1OrganizationsNamespacesGetOptions,
  v1PermissionResourceGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { v1OrganizationsNamespacesCreate } from "@/lib/client/sdk.gen";
import { normalizeFormData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

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

  const formSchema = showOrganizationSelector
    ? namespaceCreateWithOrgSchema
    : namespaceFormSchema;

  type FormValues = z.infer<typeof formSchema>;

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: "",
      description: "",
      ...(showOrganizationSelector ? { organizationId: "" } : {}),
    } as FormValues,
  });

  const selectedOrganizationId = organizationId
    ? organizationId
    : showOrganizationSelector
      ? (form.watch("organizationId" as keyof FormValues) as string | undefined)
      : undefined;

  const mutation = useFormMutation<
    { id: string },
    Options<V1OrganizationsNamespacesCreateData>,
    FormValues
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
        ]
      : [v1OrganizationsGetOptions().queryKey],
    navigateOnSuccess: organizationId
      ? {
          to: "/settings/organizations/$organizationId",
          params: { organizationId },
        }
      : {
          to: "/settings/namespaces",
        },
    transformValues: (values) => {
      const normalizedBody = normalizeFormData(
        namespaceFormSchema,
        values
      ) as NamespaceCreate;
      const orgId =
        organizationId ||
        (values as NamespaceCreateWithOrgFormValues).organizationId;
      return {
        path: {
          id: orgId,
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
        <Form {...form}>
          <form
            onSubmit={mutation.handleSubmit}
            className="flex flex-col gap-y-6"
          >
            {mutation.isError && (
              <div className="text-destructive text-sm">
                {mutation.error?.message}
              </div>
            )}

            {showOrganizationSelector && (
              <FormField
                control={form.control}
                // @ts-expect-error - organizationId field is a hack to use namespace
                //  create form from organization and namespace pages
                name="organizationId"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Organization</FormLabel>
                    <FormControl>
                      {isLoading ? (
                        <Skeleton className="h-10 w-full" />
                      ) : writableOrganizations &&
                        writableOrganizations.length > 0 ? (
                        <EntitySelect
                          options={writableOrganizations.map((org) => ({
                            value: org.id,
                            title: org.name,
                            description: org.email || org.website || undefined,
                            avatarSrc:
                              (org as { logo?: string | null }).logo || null,
                            avatarFallback: org.name,
                          }))}
                          value={field.value as string}
                          onValueChange={field.onChange}
                          placeholder="Select an organization"
                          disabled={mutation.isPending}
                        />
                      ) : (
                        <div className="text-muted-foreground text-sm">
                          No organizations available
                        </div>
                      )}
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}

            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="Enter namespace name"
                      {...field}
                      disabled={mutation.isPending}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="Enter namespace description (optional)"
                      {...field}
                      value={getDefaultValue(field.value)}
                      rows={4}
                      disabled={mutation.isPending}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

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
        </Form>
      </CardContent>
    </Card>
  );
}
