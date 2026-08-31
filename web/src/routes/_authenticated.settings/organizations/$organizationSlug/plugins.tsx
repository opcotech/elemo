import { createFileRoute } from "@tanstack/react-router";

import { OrganizationNotFound } from "@/components/organizations/organization-not-found";
import { PluginActivationManager } from "@/components/plugins/plugin-activation";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { PageHeader } from "@/components/ui/page-header";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import { Action, can } from "@/lib/auth/permissions";
import { withRouteErrors } from "@/lib/route-errors";
import { requireOrganizationAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/plugins"
)({
  beforeLoad: ({ context, params }) =>
    requireOrganizationAction(
      context.queryClient,
      params.organizationSlug,
      Action.OrganizationRead
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      requireOrganizationAction(
        context.queryClient,
        params.organizationSlug,
        Action.OrganizationRead
      )
    ),
  staticData: { breadcrumb: "Plugins" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: OrganizationNotFound,
  component: OrganizationPluginsPage,
});

function OrganizationPluginsPage() {
  const organization = Route.useLoaderData();
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Organization, organization.id)
  );
  const canManage = can(permissions, Action.PluginManage);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Plugins"
        description="Enable installed plugins for this organization. Descendant namespaces and projects inherit active plugins."
      />
      <PluginActivationManager
        scopeId={organization.id}
        scopeType="Organization"
        canManage={canManage}
      />
    </div>
  );
}
