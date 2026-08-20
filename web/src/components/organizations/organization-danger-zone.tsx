import { EntityDangerZone } from "@/components/entity-lifecycle/entity-lifecycle";
import { organizationLifecycleConfig } from "@/components/entity-lifecycle/entity-lifecycle-configs";
import { v1OrganizationDeleteMutation } from "@/lib/api/mutation-options";
import type { EffectiveActions, Organization } from "@/lib/api/types";

interface OrganizationDangerZoneProps {
  organization: Organization;
  permissions: EffectiveActions;
}

export function OrganizationDangerZone({
  organization,
  permissions,
}: OrganizationDangerZoneProps) {
  return (
    <EntityDangerZone
      entity={organization}
      context={undefined}
      permissions={permissions}
      config={organizationLifecycleConfig}
      mutationOptions={v1OrganizationDeleteMutation()}
    />
  );
}
