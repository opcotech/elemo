import type { ErrorComponentProps } from "@tanstack/react-router";
import { createFileRoute } from "@tanstack/react-router";

import { ScopedDocumentsList } from "@/components/documents/scoped-documents-list";
import { OrganizationNotFound } from "@/components/organizations";
import {
  ProjectDangerZone,
  ProjectDetailInfo,
  ProjectIssuesList,
} from "@/components/projects";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { PageHeader } from "@/components/shared/page-header";
import { isPermissionDenied } from "@/lib/api/errors";
import { entityBreadcrumb } from "@/lib/breadcrumb";
import { loadProjectDetail } from "@/lib/route-data";
import { isAccessDeniedRouteData, withRouteErrors } from "@/lib/route-errors";

function ProjectAccessDenied() {
  return (
    <div className="space-y-6">
      <PageHeader title="Access Denied" />
      <div className="text-muted-foreground">
        You do not have permission to view this project.
      </div>
    </div>
  );
}

function ProjectDetailError({ error }: ErrorComponentProps) {
  if (isPermissionDenied(error)) {
    return <ProjectAccessDenied />;
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
  notFoundComponent: OrganizationNotFound,
  component: ProjectDetailPage,
});

function ProjectDetailPage() {
  const loaderData = Route.useLoaderData();

  if (isAccessDeniedRouteData(loaderData)) {
    return <ProjectAccessDenied />;
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

      <ScopedDocumentsList
        scope="project"
        documents={project.documents || []}
        isLoading={false}
        error={null}
      />

      <ProjectIssuesList
        issues={project.issues || []}
        isLoading={false}
        error={null}
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
