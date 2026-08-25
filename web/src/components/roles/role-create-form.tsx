import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useForm } from "react-hook-form";

import type { RoleFormValues } from "./role-form-fields";
import { RoleFormFields, roleFormSchema } from "./role-form-fields";

import { FieldProvider } from "@/components/ui/field";
import { FormCard } from "@/components/ui/form-card";
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
import { showSuccessToast } from "@/lib/toast";

interface RoleCreateFormProps {
  organizationId: string;
  organizationSlug: string;
}

export function RoleCreateForm({
  organizationId,
  organizationSlug,
}: RoleCreateFormProps) {
  const navigate = useNavigate();
  const { pendingActions } = usePendingPermissions();

  const form = useForm<RoleFormValues>({
    resolver: zodResolver(roleFormSchema),
    defaultValues: {
      name: "",
      key: "",
      description: "",
      actions: pendingActions,
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
        path: { organizationRef: organizationId },
      }).queryKey,
    ],
    navigateOnSuccess: undefined,
    transformValues: (values) => {
      const normalizedBody = normalizeFormData(
        roleFormSchema,
        values
      ) as RoleCreate;
      return {
        path: { organizationRef: organizationId },
        body: {
          ...normalizedBody,
          actions: values.actions ?? pendingActions,
        },
      };
    },
    onSuccess: () => {
      showSuccessToast("Role created", "The role was created successfully");
      navigate({
        to: "/settings/organizations/$organizationSlug",
        params: { organizationSlug },
      });
    },
  });

  const handleCancel = () => {
    navigate({
      to: "/settings/organizations/$organizationSlug",
      params: { organizationSlug },
    });
  };

  return (
    <div className="flex flex-col gap-6" data-section="role-create-form">
      <FormCard
        description="Enter the role details and the inspectable actions this bundle grants."
        onSubmit={roleMutation.handleSubmit}
        onCancel={handleCancel}
        isPending={roleMutation.isPending}
        error={roleMutation.error || null}
        submitButtonText="Create Role"
      >
        <FieldProvider {...form}>
          <RoleFormFields
            control={form.control}
            isPending={roleMutation.isPending}
          />
        </FieldProvider>
      </FormCard>
    </div>
  );
}
