import { EntityDeleteDialog } from "@/components/entity-lifecycle/entity-lifecycle";
import { projectLifecycleConfig } from "@/components/entity-lifecycle/entity-lifecycle-configs";
import { v1ProjectDeleteMutation } from "@/lib/api/mutation-options";
import type { PartialProject, Project } from "@/lib/api/types";

interface ProjectDeleteDialogProps {
  project: Pick<Project | PartialProject, "id" | "name">;
  organizationId: string;
  namespaceId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
  navigateOnSuccess?: boolean;
}

export function ProjectDeleteDialog({
  project,
  organizationId,
  namespaceId,
  open,
  onOpenChange,
  onSuccess,
  navigateOnSuccess = false,
}: ProjectDeleteDialogProps) {
  return (
    <EntityDeleteDialog
      entity={project}
      context={{ organizationId, namespaceId }}
      config={projectLifecycleConfig}
      mutationOptions={v1ProjectDeleteMutation()}
      open={open}
      onOpenChange={onOpenChange}
      onSuccess={onSuccess}
      navigateOnSuccess={navigateOnSuccess}
    />
  );
}
