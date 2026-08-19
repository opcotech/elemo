import { EntityDeleteDialog } from "@/components/entity-lifecycle/entity-lifecycle";
import { teamLifecycleConfig } from "@/components/entity-lifecycle/entity-lifecycle-configs";
import { v1OrganizationTeamDeleteMutation } from "@/lib/api/mutation-options";
import type { Team } from "@/lib/api/types";

interface TeamDeleteDialogProps {
  team: Team;
  organizationId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function TeamDeleteDialog({
  team,
  organizationId,
  open,
  onOpenChange,
  onSuccess,
}: TeamDeleteDialogProps) {
  return (
    <EntityDeleteDialog
      entity={team}
      context={{ organizationId }}
      config={teamLifecycleConfig}
      mutationOptions={v1OrganizationTeamDeleteMutation()}
      open={open}
      onOpenChange={onOpenChange}
      onSuccess={onSuccess}
    />
  );
}
