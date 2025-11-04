import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";

import type {RoleFormValues} from "./role-form-fields";
import { RoleFormFields, roleFormSchema   } from "./role-form-fields";
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
import { Form } from "@/components/ui/form";
import { Separator } from "@/components/ui/separator";
import { usePendingPermissions } from "@/hooks/use-pending-permissions";
import type { RoleCreate } from "@/lib/api";
import {
  v1OrganizationRolesCreateMutation,
  v1OrganizationRolesGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { normalizeFormData } from "@/lib/forms";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

interface OrganizationRoleCreateFormProps {
  organizationId: string;
}

export function OrganizationRoleCreateForm({
  organizationId,
}: OrganizationRoleCreateFormProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
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

  const roleMutation = useMutation(v1OrganizationRolesCreateMutation());

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

  const onSubmit = async (values: RoleFormValues) => {
    const normalizedBody = normalizeFormData(
      roleFormSchema,
      values
    ) as RoleCreate;

    try {
      // Create the role first
      const data = await roleMutation.mutateAsync({
        path: {
          id: organizationId,
        },
        body: normalizedBody,
      });

      const roleId = data.id;
      setCreatedRoleId(roleId);

      // Create pending permissions if any
      if (hasPendingPermissions) {
        const { success, failed, total } = await createPermissions(roleId);

        if (failed === 0) {
          showSuccessToast(
            "Role created",
            `Role created successfully with ${success} permission(s)`
          );
        } else if (success > 0) {
          showSuccessToast(
            "Role created",
            `Role created with ${success}/${total} permission(s)`
          );
        } else {
          showSuccessToast(
            "Role created",
            "Role created but permissions failed to create"
          );
        }
        clearPermissions();
      } else {
        showSuccessToast("Role created", "Role created successfully");
      }

      // Invalidate roles list query
      queryClient.invalidateQueries({
        queryKey: v1OrganizationRolesGetOptions({
          path: { id: organizationId },
        }).queryKey,
      });
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to create role";
      showErrorToast("Failed to create role", errorMessage);
    }
  };

  const isSubmitting = roleMutation.isPending || isCreatingPermissions;

  // Show success state after role creation
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
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Create Role</CardTitle>
          <CardDescription>
            Enter the role details below and optionally add permissions.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form>
            <RoleFormFields
              isPending={isSubmitting}
              errorMessage={
                roleMutation.isError ? roleMutation.error.message : undefined
              }
              onCancel={handleCancel}
              submitButtonText={
                hasPendingPermissions
                  ? `Create Role with ${pendingPermissions.length} Permission(s)`
                  : "Create Role"
              }
              onSubmit={onSubmit}
            />
          </Form>
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
