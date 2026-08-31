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
import { requireProjectAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey/plugins"
)({
  beforeLoad: ({ context, params }) =>
    requireProjectAction(
      context.queryClient,
      params.organizationSlug,
      params.namespaceSlug,
      params.projectKey,
      Action.ProjectRead
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      requireProjectAction(
        context.queryClient,
        params.organizationSlug,
        params.namespaceSlug,
        params.projectKey,
        Action.ProjectRead
      )
    ),
  staticData: { breadcrumb: "Plugins" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: ProjectPluginsPage,
});

function ProjectPluginsPage() {
  const { project } = Route.useLoaderData();
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Project, project.id)
  );
  const canManage = can(permissions, Action.PluginManage);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Plugins"
        description="Enable installed plugins for this project."
      />
      <PluginActivationManager
        scopeId={project.id}
        scopeType="Project"
        canManage={canManage}
      />
    </div>
  );
}
