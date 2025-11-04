import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { RolePermissionAssignment } from "./role-permission-assignment";
import { RolePermissionDraft } from "./role-permission-draft";
import type { PendingPermission } from "./role-permission-draft";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import type { RoleCreate } from "@/lib/api";
import {
  v1OrganizationRolePermissionAddMutation,
  v1OrganizationRolePermissionsGetOptions,
  v1OrganizationRolesCreateMutation,
  v1OrganizationRolesGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { zRoleCreate } from "@/lib/client/zod.gen";
import {
  createFormSchema,
  getFieldValue,
  normalizeFormData,
} from "@/lib/forms";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

const roleFormSchema = createFormSchema(zRoleCreate);

type RoleFormValues = z.infer<typeof roleFormSchema>;

const defaultValues: RoleFormValues = {
  name: "",
  description: "",
};

interface OrganizationRoleCreateFormProps {
  organizationId: string;
}

export function OrganizationRoleCreateForm({
  organizationId,
}: OrganizationRoleCreateFormProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const form = useForm<RoleFormValues>({
    resolver: zodResolver(roleFormSchema),
    defaultValues,
  });

  const mutation = useMutation(v1OrganizationRolesCreateMutation());

  const onSubmit = (values: RoleFormValues) => {
    const normalizedBody = normalizeFormData(
      roleFormSchema,
      values
    ) as RoleCreate;

    mutation.mutate(
      {
        path: {
          id: organizationId,
        },
        body: normalizedBody,
      },
      {
        onSuccess: () => {
          showSuccessToast("Role created", "Role created successfully");

          // Invalidate roles list query
          queryClient.invalidateQueries({
            queryKey: v1OrganizationRolesGetOptions({
              path: { id: organizationId },
            }).queryKey,
          });

          navigate({
            to: "/settings/organizations/$organizationId",
            params: { organizationId },
          });
        },
        onError: (error) => {
          showErrorToast("Failed to create role", error.message);
        },
      }
    );
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create Role</CardTitle>
        <CardDescription>
          Enter the details below to create a new role.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex flex-col gap-y-6"
          >
            {mutation.isError && (
              <Alert variant="destructive">
                <AlertTitle>Failed to create role</AlertTitle>
                <AlertDescription>{mutation.error.message}</AlertDescription>
              </Alert>
            )}

            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Enter role name" {...field} />
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
                      placeholder="Enter role description (optional)"
                      {...field}
                      value={getFieldValue(field.value)}
                      rows={4}
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
                    <span>Creating...</span>
                  </>
                ) : (
                  "Create Role"
                )}
              </Button>
            </div>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}

export function OrganizationRoleCreateFormWithPermissions({
  organizationId,
}: OrganizationRoleCreateFormProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [createdRoleId, setCreatedRoleId] = useState<string | null>(null);
  const [pendingPermissions, setPendingPermissions] = useState<
    PendingPermission[]
  >([]);

  const form = useForm<RoleFormValues>({
    resolver: zodResolver(roleFormSchema),
    defaultValues,
  });

  const roleMutation = useMutation(v1OrganizationRolesCreateMutation());
  const permissionMutation = useMutation(
    v1OrganizationRolePermissionAddMutation()
  );

  const handleAddPermission = (permission: PendingPermission) => {
    setPendingPermissions((prev) => [...prev, permission]);
  };

  const handleRemovePermission = (index: number) => {
    setPendingPermissions((prev) => prev.filter((_, i) => i !== index));
  };

  const createPermissions = async (roleId: string) => {
    if (pendingPermissions.length === 0) {
      return { success: 0, failed: 0 };
    }

    const errors: string[] = [];
    let successCount = 0;

    // Create permissions sequentially
    for (const permission of pendingPermissions) {
      try {
        await permissionMutation.mutateAsync({
          path: {
            id: organizationId,
            role_id: roleId,
          },
          body: {
            target: permission.target,
            kind: permission.kind as
              | "read"
              | "write"
              | "create"
              | "delete"
              | "*",
          },
        });
        successCount++;
      } catch (error) {
        const errorMessage =
          error instanceof Error
            ? error.message
            : "Failed to create permission";
        errors.push(
          `${permission.resourceType}:${permission.resourceId} (${permission.kind}) - ${errorMessage}`
        );
      }
    }

    if (errors.length > 0) {
      showErrorToast("Some permissions failed to create", errors.join("; "));
    }

    return { success: successCount, failed: errors.length };
  };

  const onSubmit = (values: RoleFormValues) => {
    const normalizedBody = normalizeFormData(
      roleFormSchema,
      values
    ) as RoleCreate;

    roleMutation.mutate(
      {
        path: {
          id: organizationId,
        },
        body: normalizedBody,
      },
      {
        onSuccess: async (data) => {
          const roleId = data.id;
          setCreatedRoleId(roleId);

          // Create all pending permissions
          if (pendingPermissions.length > 0) {
            const { success, failed } = await createPermissions(roleId);

            // Invalidate permissions query to refresh the list
            queryClient.invalidateQueries({
              queryKey: v1OrganizationRolePermissionsGetOptions({
                path: {
                  id: organizationId,
                  role_id: roleId,
                },
              }).queryKey,
            });

            if (success > 0 && failed === 0) {
              showSuccessToast(
                "Role created",
                `Role created successfully with ${success} permission(s)`
              );
            } else if (success > 0) {
              showSuccessToast(
                "Role created",
                `Role created successfully with ${success} permission(s), ${failed} failed`
              );
            } else {
              showSuccessToast(
                "Role created",
                "Role created successfully, but all permissions failed"
              );
            }
          } else {
            showSuccessToast("Role created", "Role created successfully");
          }

          // Invalidate roles list query
          queryClient.invalidateQueries({
            queryKey: v1OrganizationRolesGetOptions({
              path: { id: organizationId },
            }).queryKey,
          });

          // Clear pending permissions
          setPendingPermissions([]);
        },
        onError: (error) => {
          showErrorToast("Failed to create role", error.message);
        },
      }
    );
  };

  const handleContinue = () => {
    navigate({
      to: "/settings/organizations/$organizationId",
      params: { organizationId },
    });
  };

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Create Role</CardTitle>
          <CardDescription>
            Enter the details below to create a new role.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(onSubmit)}
              className="flex flex-col gap-y-6"
            >
              {roleMutation.isError && (
                <Alert variant="destructive">
                  <AlertTitle>Failed to create role</AlertTitle>
                  <AlertDescription>
                    {roleMutation.error.message}
                  </AlertDescription>
                </Alert>
              )}

              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input placeholder="Enter role name" {...field} />
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
                        placeholder="Enter role description (optional)"
                        {...field}
                        value={getFieldValue(field.value)}
                        rows={4}
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
                  onClick={() =>
                    navigate({
                      to: "/settings/organizations/$organizationId",
                      params: { organizationId },
                    })
                  }
                  disabled={roleMutation.isPending}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={roleMutation.isPending}>
                  {roleMutation.isPending ? (
                    <>
                      <Spinner size="xs" className="mr-0.5 text-white" />
                      <span>Creating...</span>
                    </>
                  ) : (
                    "Create Role"
                  )}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>

      {!createdRoleId && (
        <RolePermissionDraft
          permissions={pendingPermissions}
          onAddPermission={handleAddPermission}
          onRemovePermission={handleRemovePermission}
        />
      )}

      {createdRoleId && (
        <>
          <RolePermissionAssignment
            organizationId={organizationId}
            roleId={createdRoleId}
          />

          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={handleContinue}>
              Continue to Organization
            </Button>
          </div>
        </>
      )}
    </div>
  );
}
