import { EntityDangerZone } from "@/components/entity-lifecycle/entity-lifecycle";
import { projectLifecycleConfig } from "@/components/entity-lifecycle/entity-lifecycle-configs";
import { v1ProjectDeleteMutation } from "@/lib/api/mutation-options";
import type { Permission, Project } from "@/lib/api/types";

interface ProjectDangerZoneProps {
  project: Project;
  permissions: Permission[];
  organizationId: string;
  namespaceId: string;
}

export function ProjectDangerZone({
  project,
  permissions,
  organizationId,
  namespaceId,
}: ProjectDangerZoneProps) {
  return (
    <EntityDangerZone
      entity={project}
      context={{ organizationId, namespaceId }}
      permissions={permissions}
      config={projectLifecycleConfig}
      mutationOptions={v1ProjectDeleteMutation()}
    />
  );
}
