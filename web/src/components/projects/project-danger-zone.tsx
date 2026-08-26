import { EntityDangerZone } from "@/components/entity-lifecycle/entity-lifecycle";
import { projectLifecycleConfig } from "@/components/entity-lifecycle/entity-lifecycle-configs";
import { v1ProjectDeleteMutation } from "@/lib/api/mutation-options";
import type { EffectiveActions, Project } from "@/lib/api/types";

interface ProjectDangerZoneProps {
  project: Project;
  permissions: EffectiveActions;
  organizationId: string;
  organizationSlug: string;
  namespaceId: string;
  namespaceSlug: string;
}

export function ProjectDangerZone({
  project,
  permissions,
  organizationId,
  organizationSlug,
  namespaceId,
  namespaceSlug,
}: ProjectDangerZoneProps) {
  return (
    <EntityDangerZone
      entity={project}
      context={{ organizationId, organizationSlug, namespaceId, namespaceSlug }}
      permissions={permissions}
      config={projectLifecycleConfig}
      mutationOptions={v1ProjectDeleteMutation()}
    />
  );
}
