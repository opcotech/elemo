import { createFileRoute } from "@tanstack/react-router";

import { PluginCatalog } from "@/components/plugins/plugin-catalog";
import { PageHeader } from "@/components/ui/page-header";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import { Action, can } from "@/lib/auth/permissions";
import { requirePermissionBeforeLoad } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/_authenticated/settings/plugins")({
  beforeLoad: ({ context }) =>
    requirePermissionBeforeLoad({
      queryClient: context.queryClient,
      resourceType: ResourceType.Installation,
      action: Action.PluginInstall,
    }),
  staticData: { breadcrumb: "Plugins" },
  component: InstancePluginsPage,
});

function InstancePluginsPage() {
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Installation)
  );
  const canInstall = can(permissions, Action.PluginInstall);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Plugins"
        description="Install and uninstall plugin packages for this Elemo instance. Enabling a plugin happens per organization, namespace, or project."
      />
      <PluginCatalog canInstall={canInstall} />
    </div>
  );
}
