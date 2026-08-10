import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";

import { OrganizationNotFound } from "@/components/organizations";
import { ProjectCreateForm } from "@/components/projects";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { PageHeader } from "@/components/shared/page-header";
import {
  v1NamespaceGetOptions,
  v1OrganizationGetOptions,
} from "@/lib/api/query-options";
import { ResourceType } from "@/lib/auth/permissions";
import { requirePermissionBeforeLoad } from "@/lib/auth/require-auth";
import { loadNamespaceHierarchy } from "@/lib/route-data";
import { withRouteErrors } from "@/lib/route-errors";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationId/namespaces/$namespaceId/projects/new"
)({
  beforeLoad: ({ context, params }) =>
    requirePermissionBeforeLoad({
      queryClient: context.queryClient,
      resourceType: ResourceType.Namespace,
      permissionKind: "write",
      resourceId: params.namespaceId,
    }),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      loadNamespaceHierarchy(
        context.queryClient,
        params.organizationId,
        params.namespaceId
      )
    ),
  staticData: { breadcrumb: "Create project" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: OrganizationNotFound,
  component: OrganizationNamespaceProjectCreatePage,
});

function OrganizationNamespaceProjectCreatePage() {
  const { organizationId, namespaceId } = Route.useParams();
  useSuspenseQuery(
    v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

  const { data: namespace } = useSuspenseQuery(
    v1NamespaceGetOptions({
      path: {
        id: namespaceId,
      },
    })
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title="Create Project"
        description={`Create a new project in ${namespace.name}.`}
      />

      <ProjectCreateForm
        organizationId={organizationId}
        namespaceId={namespaceId}
      />
    </div>
  );
}
