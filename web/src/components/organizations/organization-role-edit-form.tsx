import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { RoleMemberAssignment } from "./role-member-assignment";
import { RolePermissionAssignment } from "./role-permission-assignment";

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
import type { Role } from "@/lib/api";
import {
  v1OrganizationRoleGetOptions,
  v1OrganizationRoleUpdateMutation,
  v1OrganizationRolesGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { zRolePatch } from "@/lib/client/zod.gen";
import {
  createFormSchema,
  getFieldValue,
  normalizePatchData,
} from "@/lib/forms";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

const roleEditFormSchema = createFormSchema(zRolePatch);

type RoleEditFormValues = z.infer<typeof roleEditFormSchema>;

interface OrganizationRoleEditFormProps {
  role: Role;
  organizationId: string;
  roleId: string;
}

export function OrganizationRoleEditForm({
  role,
  organizationId,
  roleId,
}: OrganizationRoleEditFormProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const form = useForm<RoleEditFormValues>({
    resolver: zodResolver(roleEditFormSchema),
    defaultValues: {
      name: role.name,
      description: getFieldValue(role.description),
    },
  });

  // Update form when role data changes (but not while user is editing)
  useEffect(() => {
    // Only reset if the form is not being actively edited (isDirty check)
    if (!form.formState.isDirty) {
      form.reset({
        name: role.name,
        description: getFieldValue(role.description),
      });
    }
  }, [role.name, role.description, form]);

  const mutation = useMutation(v1OrganizationRoleUpdateMutation());

  const onSubmit = (values: RoleEditFormValues) => {
    // Normalize patch data: converts empty strings to null for cleared optional fields
    const normalizedBody = normalizePatchData(roleEditFormSchema, values, {
      name: role.name,
      description: role.description,
    });

    mutation.mutate(
      {
        path: {
          id: organizationId,
          role_id: roleId,
        },
        body: normalizedBody,
      },
      {
        onSuccess: () => {
          showSuccessToast("Role updated", "Role updated successfully");

          // Invalidate queries to refresh the list and detail views
          queryClient.invalidateQueries({
            queryKey: v1OrganizationRolesGetOptions({
              path: { id: organizationId },
            }).queryKey,
          });
          queryClient.invalidateQueries({
            queryKey: v1OrganizationRoleGetOptions({
              path: {
                id: organizationId,
                role_id: roleId,
              },
            }).queryKey,
          });

          navigate({
            to: "/settings/organizations/$organizationId",
            params: { organizationId },
          });
        },
        onError: (error) => {
          showErrorToast("Failed to update role", error.message);
        },
      }
    );
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Edit Role</CardTitle>
        <CardDescription>Update the role details below.</CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex flex-col gap-y-6"
          >
            {mutation.isError && (
              <Alert variant="destructive">
                <AlertTitle>Failed to update role</AlertTitle>
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
                    <span>Saving...</span>
                  </>
                ) : (
                  "Save Changes"
                )}
              </Button>
            </div>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}

export function OrganizationRoleEditFormWithPermissions({
  role,
  organizationId,
  roleId,
}: OrganizationRoleEditFormProps) {
  return (
    <div className="flex flex-col gap-6">
      <OrganizationRoleEditForm
        role={role}
        organizationId={organizationId}
        roleId={roleId}
      />
      <RoleMemberAssignment
        organizationId={organizationId}
        roleId={roleId}
        roleName={role.name}
      />
      <RolePermissionAssignment
        organizationId={organizationId}
        roleId={roleId}
      />
    </div>
  );
}
