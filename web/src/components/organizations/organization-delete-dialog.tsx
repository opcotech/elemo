import { EntityDeleteDialog } from "@/components/entity-lifecycle/entity-lifecycle";
import { organizationLifecycleConfig } from "@/components/entity-lifecycle/entity-lifecycle-configs";
import { v1OrganizationDeleteMutation } from "@/lib/api/mutation-options";
import type { Organization } from "@/lib/api/types";

interface OrganizationDeleteDialogProps {
  organization: Organization;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function OrganizationDeleteDialog({
  organization,
  open,
  onOpenChange,
  onSuccess,
}: OrganizationDeleteDialogProps) {
  return (
    <EntityDeleteDialog
      entity={organization}
      context={undefined}
      config={organizationLifecycleConfig}
      mutationOptions={v1OrganizationDeleteMutation()}
      open={open}
      onOpenChange={onOpenChange}
      onSuccess={onSuccess}
      navigateOnSuccess
    />
  );
}
