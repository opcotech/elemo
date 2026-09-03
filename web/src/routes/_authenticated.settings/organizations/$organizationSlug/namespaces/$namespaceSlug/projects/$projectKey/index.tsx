import type { ErrorComponentProps } from "@tanstack/react-router";
import { createFileRoute } from "@tanstack/react-router";

import { PluginSlot } from "@/components/plugins/plugin-slot";
import { ProjectDangerZone } from "@/components/projects/project-danger-zone";
import { ProjectDetailInfo } from "@/components/projects/project-detail-info";
import {
  SettingsAccessDenied,
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { PageHeader } from "@/components/ui/page-header";
import { isPermissionDenied } from "@/lib/api/errors";
import { entityBreadcrumb } from "@/lib/breadcrumb";
import { loadProjectDetail } from "@/lib/route-data";
import { isAccessDeniedRouteData, withRouteErrors } from "@/lib/route-errors";

function ProjectDetailError({ error }: ErrorComponentProps) {
  if (isPermissionDenied(error)) {
    return <SettingsAccessDenied resource="project" />;
  }
  return <SettingsEntityDetailError />;
}

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey/"
)({
  loader: ({ context, params }) =>
    withRouteErrors(
      () =>
        loadProjectDetail(
          context.queryClient,
          params.organizationSlug,
          params.namespaceSlug,
          params.projectKey
        ),
      "data"
    ),
  staticData: {
    breadcrumb: (data) => {
      if (isAccessDeniedRouteData(data)) {
        return "Project";
      }
      return entityBreadcrumb(data, "project", "Project");
    },
  },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: ProjectDetailError,
  notFoundComponent: SettingsNotFound,
  component: ProjectDetailPage,
});

function ProjectDetailPage() {
  const loaderData = Route.useLoaderData();

  if (isAccessDeniedRouteData(loaderData)) {
    return <SettingsAccessDenied resource="project" />;
  }

  const { project, namespace, organization, permissions } = loaderData;

  return (
    <div className="space-y-6">
      <PageHeader title={project.name} />

      <ProjectDetailInfo
        project={project}
        organizationSlug={organization.slug}
        namespaceSlug={namespace.slug}
        namespaceName={namespace.name}
      />

      <PluginSlot
        name="project.settings"
        context={{
          organizationId: organization.id,
          organizationSlug: organization.slug,
          namespaceId: namespace.id,
          namespaceSlug: namespace.slug,
          projectId: project.id,
          projectKey: project.key,
        }}
      />

      <ProjectDangerZone
        project={project}
        permissions={permissions}
        organizationId={organization.id}
        organizationSlug={organization.slug}
        namespaceId={namespace.id}
        namespaceSlug={namespace.slug}
      />
    </div>
  );
}
