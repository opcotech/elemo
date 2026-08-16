import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { RoleFormFields, roleFormSchema } from "./role-form-fields";
import { RoleMemberAssignment } from "./role-member-assignment";
import { RolePermissionAssignment } from "./role-permission-assignment";

import { FieldProvider } from "@/components/ui/field";
import { FormCard } from "@/components/ui/form-card";
import { useFormMutation } from "@/hooks/use-form-mutation";
import {
  v1OrganizationRoleGetOptions,
  v1OrganizationRolesGetOptions,
} from "@/lib/api/query-options";
import { v1OrganizationRoleUpdate } from "@/lib/api/sdk";
import type {
  Options,
  Role,
  V1OrganizationRoleUpdateData,
} from "@/lib/api/types";
import { zRolePatch } from "@/lib/client/zod.gen";
import { createFormSchema, normalizePatchData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

const roleEditFormSchema = createFormSchema(
  zRolePatch.extend({
    name: roleFormSchema.shape.name,
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
    navigateOnSuccess: (navigateTo) =>
      navigateTo({
        to: "/settings/organizations/$organizationId",
        params: { organizationId },
      }),
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
      <FieldProvider {...form}>
        <RoleFormFields control={form.control} isPending={mutation.isPending} />
      </FieldProvider>
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
