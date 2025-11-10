import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { useForm } from "react-hook-form";

import type { RoleFormValues } from "./role-form-fields";
import { RoleFormFields, roleFormSchema } from "./role-form-fields";
import { RolePermissionAssignment } from "./role-permission-assignment";
import { RolePermissionDraft } from "./role-permission-draft";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Spinner } from "@/components/ui/spinner";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { usePendingPermissions } from "@/hooks/use-pending-permissions";
import type {
  Options,
  RoleCreate,
  V1OrganizationRolesCreateData,
} from "@/lib/api";
import { v1OrganizationRolesGetOptions } from "@/lib/client/@tanstack/react-query.gen";
import { v1OrganizationRolesCreate } from "@/lib/client/sdk.gen";
import { normalizeFormData } from "@/lib/forms";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

interface RoleCreateFormProps {
  organizationId: string;
}

export function RoleCreateForm({ organizationId }: RoleCreateFormProps) {
  const navigate = useNavigate();
  const [createdRoleId, setCreatedRoleId] = useState<string | null>(null);

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
    successMessage: undefined, // We show custom success message in onSuccess
    errorMessagePrefix: "Failed to create role",
    queryKeysToInvalidate: [
      v1OrganizationRolesGetOptions({
        path: { id: organizationId },
      }).queryKey,
    ],
    // Don't navigate automatically - we'll handle it after permissions
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
      const roleId = data.id;
      setCreatedRoleId(roleId);

      const { success, failed } = await createPermissions(roleId);

      // Show detailed toast (will briefly show two toasts, but detailed one is more informative)
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

  const handleContinue = () => {
    navigate({
      to: "/settings/organizations/$organizationId",
      params: { organizationId },
    });
  };

  const isSubmitting = roleMutation.isPending || isCreatingPermissions;

  if (createdRoleId) {
    return (
      <div className="flex flex-col gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Role Created Successfully</CardTitle>
            <CardDescription>
              You can now manage permissions or return to the organization.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex justify-end gap-2">
              <Button variant="default" onClick={handleContinue}>
                Back to Organization
              </Button>
            </div>
          </CardContent>
        </Card>

        <RolePermissionAssignment
          organizationId={organizationId}
          roleId={createdRoleId}
        />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6" data-section="role-create-form">
      <Card>
        <CardHeader>
          <CardDescription>
            Enter the role details below and optionally add permissions.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <RoleFormFields
            form={form}
            isPending={isSubmitting}
            errorMessage={
              roleMutation.isError ? roleMutation.error?.message : undefined
            }
            onCancel={handleCancel}
            submitButtonText={
              hasPendingPermissions
                ? `Create Role with ${pendingPermissions.length} Permission(s)`
                : "Create Role"
            }
            onSubmit={roleMutation.handleSubmit}
          />
        </CardContent>
      </Card>

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

export function RoleCreateFormWithPermissions({
  organizationId,
}: RoleCreateFormProps) {
  return <RoleCreateForm organizationId={organizationId} />;
}
