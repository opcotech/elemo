import { EntityDeleteDialog } from "@/components/entity-lifecycle/entity-lifecycle";
import { roleLifecycleConfig } from "@/components/entity-lifecycle/entity-lifecycle-configs";
import { v1OrganizationRoleDeleteMutation } from "@/lib/api/mutation-options";
import type { Role } from "@/lib/api/types";

interface RoleDeleteDialogProps {
  role: Role;
  organizationId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function RoleDeleteDialog({
  role,
  organizationId,
  open,
  onOpenChange,
  onSuccess,
}: RoleDeleteDialogProps) {
  return (
    <EntityDeleteDialog
      entity={role}
      context={{ organizationId }}
      config={roleLifecycleConfig}
      mutationOptions={v1OrganizationRoleDeleteMutation()}
      open={open}
      onOpenChange={onOpenChange}
      onSuccess={onSuccess}
    />
  );
}
