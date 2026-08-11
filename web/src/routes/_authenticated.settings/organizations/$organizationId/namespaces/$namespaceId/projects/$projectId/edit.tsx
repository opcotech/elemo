import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";

import { OrganizationNotFound } from "@/components/organizations";
import { ProjectEditForm } from "@/components/projects";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { PageHeader } from "@/components/shared/page-header";
import {
  v1NamespaceGetOptions,
  v1OrganizationGetOptions,
  v1ProjectGetOptions,
} from "@/lib/api/query-options";
import { ResourceType } from "@/lib/auth/permissions";
import { requirePermissionBeforeLoad } from "@/lib/auth/require-auth";
import { loadProjectHierarchy } from "@/lib/route-data";
import { withRouteErrors } from "@/lib/route-errors";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId/edit"
)({
  beforeLoad: ({ context, params }) =>
    requirePermissionBeforeLoad({
      queryClient: context.queryClient,
      resourceType: ResourceType.Project,
      permissionKind: "write",
      resourceId: params.projectId,
    }),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      loadProjectHierarchy(
        context.queryClient,
        params.organizationId,
        params.namespaceId,
        params.projectId
      )
    ),
  staticData: { breadcrumb: "Edit project" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: OrganizationNotFound,
  component: ProjectEditPage,
});

function ProjectEditPage() {
  const { organizationId, namespaceId, projectId } = Route.useParams();
  useSuspenseQuery(
    v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

  useSuspenseQuery(
    v1NamespaceGetOptions({
      path: {
        id: namespaceId,
      },
    })
  );

  const { data: project } = useSuspenseQuery(
    v1ProjectGetOptions({
      path: {
        id: projectId,
      },
    })
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title="Edit Project"
        description="Update the project details below."
      />

      <ProjectEditForm
        project={project}
        organizationId={organizationId}
        namespaceId={namespaceId}
      />
    </div>
  );
}
