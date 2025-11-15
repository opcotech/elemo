import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { RoleMemberAssignment } from "./role-member-assignment";
import { RolePermissionAssignment } from "./role-permission-assignment";

import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { FormCard } from "@/components/ui/form-card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useFormMutation } from "@/hooks/use-form-mutation";
import type { Options, Role, V1OrganizationRoleUpdateData } from "@/lib/api";
import {
  v1OrganizationRoleGetOptions,
  v1OrganizationRolesGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { v1OrganizationRoleUpdate } from "@/lib/client/sdk.gen";
import { zRoleCreate, zRolePatch } from "@/lib/client/zod.gen";
import { createFormSchema, normalizePatchData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

const roleEditFormSchema = createFormSchema(
  zRolePatch.extend({
    name: zRoleCreate.def.shape.name,
  })
);

type RoleEditFormValues = z.infer<typeof roleEditFormSchema>;

interface RoleEditFormProps {
  role: Role;
  organizationId: string;
  roleId: string;
}

export function RoleEditForm({
  role,
  organizationId,
  roleId,
}: RoleEditFormProps) {
  const navigate = useNavigate();

  const form = useForm<RoleEditFormValues>({
    resolver: zodResolver(roleEditFormSchema),
    defaultValues: {
      name: role.name,
      description: getDefaultValue(role.description),
    },
  });

  useEffect(() => {
    if (!form.formState.isDirty) {
      form.reset({
        name: role.name,
        description: getDefaultValue(role.description),
      });
    }
  }, [role.name, role.description, form]);

  const mutation = useFormMutation<
    Role,
    Options<V1OrganizationRoleUpdateData>,
    RoleEditFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1OrganizationRoleUpdate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Role updated",
    errorMessagePrefix: "Failed to update role",
    queryKeysToInvalidate: [
      v1OrganizationRolesGetOptions({
        path: { id: organizationId },
      }).queryKey,
      v1OrganizationRoleGetOptions({
        path: {
          id: organizationId,
          role_id: roleId,
        },
      }).queryKey,
    ],
    navigateOnSuccess: {
      to: "/settings/organizations/$organizationId",
      params: { organizationId },
    },
    transformValues: (values) => {
      const normalizedBody = normalizePatchData(roleEditFormSchema, values, {
        name: role.name,
        description: role.description,
      });
      return {
        path: {
          id: organizationId,
          role_id: roleId,
        },
        body: normalizedBody,
      };
    },
  });

  return (
    <FormCard
      data-section="role-edit-form"
      onSubmit={mutation.handleSubmit}
      onCancel={() =>
        navigate({
          to: "/settings/organizations/$organizationId",
          params: { organizationId },
        })
      }
      isPending={mutation.isPending}
      error={mutation.error || null}
      submitButtonText="Save Changes"
      description="Update the role details below."
    >
      <Form {...form}>
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
                  value={getDefaultValue(field.value)}
                  rows={4}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </Form>
    </FormCard>
  );
}

export function RoleEditFormWithPermissions({
  role,
  organizationId,
  roleId,
}: RoleEditFormProps) {
  return (
    <div className="flex flex-col gap-6">
      <RoleEditForm
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
