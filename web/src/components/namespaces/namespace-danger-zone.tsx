import { EntityDangerZone } from "@/components/entity-lifecycle/entity-lifecycle";
import { namespaceLifecycleConfig } from "@/components/entity-lifecycle/entity-lifecycle-configs";
import { v1NamespaceDeleteMutation } from "@/lib/api/mutation-options";
import type { EffectiveActions, Namespace } from "@/lib/api/types";

interface NamespaceDangerZoneProps {
  namespace: Namespace;
  permissions: EffectiveActions;
  organizationId: string;
  organizationSlug: string;
}

export function NamespaceDangerZone({
  namespace,
  permissions,
  organizationId,
  organizationSlug,
}: NamespaceDangerZoneProps) {
  return (
    <EntityDangerZone
      entity={namespace}
      context={{ organizationId, organizationSlug }}
      permissions={permissions}
      config={namespaceLifecycleConfig}
      mutationOptions={v1NamespaceDeleteMutation()}
    />
  );
}
