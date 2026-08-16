import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useForm } from "react-hook-form";

import type { RoleFormValues } from "./role-form-fields";
import { RoleFormFields, roleFormSchema } from "./role-form-fields";
import { RolePermissionDraft } from "./role-permission-draft";

import { Card, CardContent } from "@/components/ui/card";
import { FieldProvider } from "@/components/ui/field";
import { FormCard } from "@/components/ui/form-card";
import { Separator } from "@/components/ui/separator";
import { Spinner } from "@/components/ui/spinner";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { usePendingPermissions } from "@/hooks/use-pending-permissions";
import { v1OrganizationRolesGetOptions } from "@/lib/api/query-options";
import { v1OrganizationRolesCreate } from "@/lib/api/sdk";
import type {
  Options,
  RoleCreate,
  V1OrganizationRolesCreateData,
} from "@/lib/api/types";
import { normalizeFormData } from "@/lib/forms";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

interface RoleCreateFormProps {
  organizationId: string;
}

export function RoleCreateForm({ organizationId }: RoleCreateFormProps) {
  const navigate = useNavigate();

  const {
    pendingPermissions,
    addPermission,
    removePermission,
    clearPermissions,
    createPermissions,
    isCreatingPermissions,
    hasPendingPermissions,
  } = usePendingPermissions({ organizationId });

  const form = useForm<RoleFormValues>({
    resolver: zodResolver(roleFormSchema),
    defaultValues: {
      name: "",
      description: "",
    },
  });

  const roleMutation = useFormMutation<
    { id: string },
    Options<V1OrganizationRolesCreateData>,
    RoleFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1OrganizationRolesCreate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: undefined,
    errorMessagePrefix: "Failed to create role",
    queryKeysToInvalidate: [
      v1OrganizationRolesGetOptions({
        path: { id: organizationId },
      }).queryKey,
    ],
    navigateOnSuccess: undefined,
    transformValues: (values) => {
      const normalizedBody = normalizeFormData(
        roleFormSchema,
        values
      ) as RoleCreate;
      return {
        path: {
          id: organizationId,
        },
        body: normalizedBody,
      };
    },
    onSuccess: async (data) => {
      const { success, failed } = await createPermissions(data.id);

      if (failed === 0) {
        showSuccessToast("Role created", `The role was created successfully`);
      } else if (success > 0) {
        showErrorToast(
          "Failed to assign permissions",
          `The role was created successfully, but failed to assign ${failed} permission(s)`
        );
      } else {
        showErrorToast(
          "Failed to assign permissions",
          `The role was created successfully, but failed to assign any permissions`
        );
      }

      if (hasPendingPermissions) {
        clearPermissions();
      }

      navigate({
        to: "/settings/organizations/$organizationId",
        params: { organizationId },
      });
    },
  });

  const handleCancel = () => {
    navigate({
      to: "/settings/organizations/$organizationId",
      params: { organizationId },
    });
  };

  const isSubmitting = roleMutation.isPending || isCreatingPermissions;

  return (
    <div className="flex flex-col gap-6" data-section="role-create-form">
      <FormCard
        description="Enter the role details below and optionally add permissions."
        onSubmit={roleMutation.handleSubmit}
        onCancel={handleCancel}
        isPending={isSubmitting}
        error={roleMutation.error || null}
        submitButtonText={
          hasPendingPermissions
            ? `Create Role with ${pendingPermissions.length} Permission(s)`
            : "Create Role"
        }
      >
        <FieldProvider {...form}>
          <RoleFormFields control={form.control} isPending={isSubmitting} />
        </FieldProvider>
      </FormCard>

      {!isSubmitting && (
        <>
          <Separator />
          <RolePermissionDraft
            permissions={pendingPermissions}
            onAddPermission={addPermission}
            onRemovePermission={removePermission}
          />
        </>
      )}

      {isCreatingPermissions && (
        <Card>
          <CardContent className="py-6">
            <div className="text-muted-foreground flex items-center justify-center gap-2 text-sm">
              <Spinner size="sm" />
              <span>Creating permissions...</span>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
