import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { RoleFormFields, roleFormSchema } from "./role-form-fields";

import { FieldProvider } from "@/components/ui/field";
import { FormCard } from "@/components/ui/form-card";
import { useFormMutation } from "@/hooks/use-form-mutation";
import {
  v1OrganizationRoleGetOptions,
  v1OrganizationRolesGetOptions,
} from "@/lib/api/query-options";
import { zRolePatch } from "@/lib/api/schemas";
import { v1OrganizationRoleUpdate } from "@/lib/api/sdk";
import type {
  Options,
  Role,
  V1OrganizationRoleUpdateData,
} from "@/lib/api/types";
import { createFormSchema, normalizePatchData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

const roleEditFormSchema = createFormSchema(
  zRolePatch.extend({
    name: roleFormSchema.shape.name,
    key: roleFormSchema.shape.key,
  })
);

type RoleEditFormValues = z.infer<typeof roleEditFormSchema>;

interface RoleEditFormProps {
  role: Role;
  organizationId: string;
  organizationSlug: string;
  roleId: string;
}

export function RoleEditForm({
  role,
  organizationId,
  organizationSlug,
  roleId,
}: RoleEditFormProps) {
  const navigate = useNavigate();

  const form = useForm<RoleEditFormValues>({
    resolver: zodResolver(roleEditFormSchema),
    defaultValues: {
      name: role.name,
      key: role.key,
      description: getDefaultValue(role.description),
      actions: role.actions ?? [],
    },
  });

  useEffect(() => {
    if (!form.formState.isDirty) {
      form.reset({
        name: role.name,
        key: role.key,
        description: getDefaultValue(role.description),
        actions: role.actions ?? [],
      });
    }
  }, [role.name, role.key, role.description, role.actions, form]);

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
        path: { organizationRef: organizationId },
      }).queryKey,
      v1OrganizationRoleGetOptions({
        path: {
          organizationRef: organizationId,
          role_id: roleId,
        },
      }).queryKey,
    ],
    navigateOnSuccess: (navigateTo) =>
      navigateTo({
        to: "/settings/organizations/$organizationSlug",
        params: { organizationSlug },
      }),
    transformValues: (values) => {
      const { key: _key, ...patchValues } = values; // eslint-disable-line @typescript-eslint/no-unused-vars
      const normalizedBody = normalizePatchData(
        roleEditFormSchema,
        patchValues,
        {
          name: role.name,
          description: role.description,
          actions: role.actions,
        }
      );
      return {
        path: {
          organizationRef: organizationId,
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
          to: "/settings/organizations/$organizationSlug",
          params: { organizationSlug },
        })
      }
      isPending={mutation.isPending}
      error={mutation.error || null}
      submitButtonText="Save Changes"
      description="Update the role name, description, and bundled actions."
    >
      <FieldProvider {...form}>
        <RoleFormFields
          control={form.control}
          isPending={mutation.isPending}
          keyDisabled
        />
      </FieldProvider>
    </FormCard>
  );
}

export function RoleEditFormWithPermissions(props: RoleEditFormProps) {
  return <RoleEditForm {...props} />;
}
