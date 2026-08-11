import { EntityDangerZone } from "@/components/entity-lifecycle/entity-lifecycle";
import { namespaceLifecycleConfig } from "@/components/entity-lifecycle/entity-lifecycle-configs";
import { v1NamespaceDeleteMutation } from "@/lib/api/mutation-options";
import type { Namespace, Permission } from "@/lib/api/types";

interface NamespaceDangerZoneProps {
  namespace: Namespace;
  permissions: Permission[];
  organizationId: string;
}

export function NamespaceDangerZone({
  namespace,
  permissions,
  organizationId,
}: NamespaceDangerZoneProps) {
  return (
    <EntityDangerZone
      entity={namespace}
      context={{ organizationId }}
      permissions={permissions}
      config={namespaceLifecycleConfig}
      mutationOptions={v1NamespaceDeleteMutation()}
    />
  );
}
