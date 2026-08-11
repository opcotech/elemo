import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";

import {
  organizationScopedResourceType,
  permissionFormSchema,
} from "./permission-form-schema";
import type { PermissionFormValues } from "./permission-form-schema";

import { DialogForm } from "@/components/ui/dialog-form";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { ResourceType } from "@/hooks/use-permissions";
import { withResourceType } from "@/hooks/use-permissions";
import { v1OrganizationRolePermissionAddMutation } from "@/lib/api/mutation-options";
import { v1OrganizationRolePermissionsGetOptions } from "@/lib/api/query-options";
import { zPermissionKind } from "@/lib/client/zod.gen";
import { runMutationSuccessWorkflow } from "@/lib/mutation-workflow";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

interface RolePermissionAddDialogProps {
  organizationId: string;
  roleId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void | Promise<void>;
}

export function RolePermissionAddDialog({
  organizationId,
  roleId,
  open,
  onOpenChange,
  onSuccess,
}: RolePermissionAddDialogProps) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const [error, setError] = useState<Error | null>(null);

  const form = useForm<PermissionFormValues>({
    resolver: zodResolver(permissionFormSchema),
    defaultValues: {
      resourceType: organizationScopedResourceType.enum.Organization,
      resourceId: "",
      kind: zPermissionKind.enum.read,
    },
  });

  const rolePermissionsQueryKey = v1OrganizationRolePermissionsGetOptions({
    path: {
      id: organizationId,
      role_id: roleId,
    },
  }).queryKey;
  const mutation = useMutation({
    ...v1OrganizationRolePermissionAddMutation(),
    onSuccess: () =>
      runMutationSuccessWorkflow({
        invalidateQueries: [
          () =>
            queryClient.invalidateQueries({
              queryKey: rolePermissionsQueryKey,
            }),
        ],
        invalidateRouter: () => router.invalidate(),
        callbacks: [
          async () => {
            setError(null);
            showSuccessToast(
              "Permission added",
              "Permission added successfully"
            );
            form.reset();
            onOpenChange(false);
            await onSuccess?.();
          },
        ],
      }),
    onError: (err) => {
      setError(new Error(err.message));
      showErrorToast("Failed to add permission", err.message);
    },
  });

  useEffect(() => {
    if (open) {
      // Clear error when dialog opens
      setError(null);
    }
  }, [open]);

  const onSubmit = (values: PermissionFormValues) => {
    if (!organizationScopedResourceType.options.includes(values.resourceType)) {
      showErrorToast(
        "Invalid resource type",
        "Only organization-scoped resources can be assigned to roles"
      );
      return;
    }

    // Clear previous error when submitting again
    setError(null);

    const target = withResourceType(
      values.resourceType as ResourceType,
      values.resourceId
    );

    mutation.mutate({
      path: {
        id: organizationId,
        role_id: roleId,
      },
      body: {
        target,
        kind: values.kind,
      },
    });
  };

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Add Permission"
      onSubmit={form.handleSubmit(onSubmit)}
      isPending={mutation.isPending}
      error={error}
      submitButtonText="Add Permission"
      onReset={() => form.reset()}
      className="sm:max-w-[600px]"
    >
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <ControlledField
          control={form.control}
          name="resourceType"
          render={({ field }) => (
            <Field>
              <FieldLabel>Resource Type</FieldLabel>
              <Select
                value={field.value}
                onValueChange={field.onChange}
                items={Object.fromEntries(
                  organizationScopedResourceType.options.map((type) => [
                    type,
                    type,
                  ])
                )}
              >
                <FieldControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select resource type" />
                  </SelectTrigger>
                </FieldControl>
                <SelectContent>
                  {organizationScopedResourceType.options.map((type) => (
                    <SelectItem key={type} value={type}>
                      {type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FieldError />
            </Field>
          )}
        />

        <ControlledField
          control={form.control}
          name="resourceId"
          render={({ field }) => (
            <Field>
              <FieldLabel>Resource ID</FieldLabel>
              <FieldControl>
                <Input placeholder="Enter resource ID" {...field} />
              </FieldControl>
              <FieldError />
            </Field>
          )}
        />

        <ControlledField
          control={form.control}
          name="kind"
          render={({ field }) => (
            <Field>
              <FieldLabel>Permission Kind</FieldLabel>
              <Select
                value={field.value}
                onValueChange={field.onChange}
                items={Object.fromEntries(
                  zPermissionKind.options.map((kind) => [kind, kind])
                )}
              >
                <FieldControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select permission kind" />
                  </SelectTrigger>
                </FieldControl>
                <SelectContent>
                  {zPermissionKind.options.map((kind) => (
                    <SelectItem key={kind} value={kind}>
                      {kind}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FieldError />
            </Field>
          )}
        />
      </div>
    </DialogForm>
  );
}
