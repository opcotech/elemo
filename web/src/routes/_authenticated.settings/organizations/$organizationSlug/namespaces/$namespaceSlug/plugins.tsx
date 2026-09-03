import { createFileRoute } from "@tanstack/react-router";

import { PluginActivationManager } from "@/components/plugins/plugin-activation";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { PageHeader } from "@/components/ui/page-header";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import { Action, can } from "@/lib/auth/permissions";
import { withRouteErrors } from "@/lib/route-errors";
import { requireNamespaceAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/plugins"
)({
  beforeLoad: ({ context, params }) =>
    requireNamespaceAction(
      context.queryClient,
      params.organizationSlug,
      params.namespaceSlug,
      Action.NamespaceRead
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      requireNamespaceAction(
        context.queryClient,
        params.organizationSlug,
        params.namespaceSlug,
        Action.NamespaceRead
      )
    ),
  staticData: { breadcrumb: "Plugins" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: NamespacePluginsPage,
});

function NamespacePluginsPage() {
  const { namespace } = Route.useLoaderData();
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Namespace, namespace.id)
  );
  const canManage = can(permissions, Action.PluginManage);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Plugins"
        description="Enable installed plugins for this namespace."
      />
      <PluginActivationManager
        scopeId={namespace.id}
        scopeType="Namespace"
        canManage={canManage}
      />
    </div>
  );
}
