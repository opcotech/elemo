import type { ErrorComponentProps } from "@tanstack/react-router";
import { createFileRoute } from "@tanstack/react-router";

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
  "/_authenticated/settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId/"
)({
  loader: ({ context, params }) =>
    withRouteErrors(
      () =>
        loadProjectDetail(
          context.queryClient,
          params.organizationId,
          params.namespaceId,
          params.projectId
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

  const { organizationId, namespaceId } = Route.useParams();
  const { project, namespace, permissions } = loaderData;

  return (
    <div className="space-y-6">
      <PageHeader title={project.name} />

      <ProjectDetailInfo
        project={project}
        organizationId={organizationId}
        namespaceId={namespaceId}
        namespaceName={namespace.name}
      />

      <ProjectDangerZone
        project={project}
        permissions={permissions}
        organizationId={organizationId}
        namespaceId={namespaceId}
      />
    </div>
  );
}
