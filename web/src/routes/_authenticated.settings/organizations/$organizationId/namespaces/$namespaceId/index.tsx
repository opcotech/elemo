import { createFileRoute } from "@tanstack/react-router";

import { ScopedDocumentsList } from "@/components/documents/scoped-documents-list";
import {
  NamespaceDangerZone,
  NamespaceDetailInfo,
  NamespaceProjectsList,
} from "@/components/namespaces";
import { OrganizationNotFound } from "@/components/organizations";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { PageHeader } from "@/components/shared/page-header";
import { entityBreadcrumb } from "@/lib/breadcrumb";
import { loadNamespaceDetail } from "@/lib/route-data";
import { isAccessDeniedRouteData, withRouteErrors } from "@/lib/route-errors";

function NamespaceAccessDenied() {
  return (
    <div className="space-y-6">
      <PageHeader title="Access Denied" />
      <div className="text-muted-foreground">
        You do not have permission to view this namespace.
      </div>
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationId/namespaces/$namespaceId/"
)({
  loader: ({ context, params }) =>
    withRouteErrors(
      () =>
        loadNamespaceDetail(
          context.queryClient,
          params.organizationId,
          params.namespaceId
        ),
      "data"
    ),
  staticData: {
    breadcrumb: (data) => entityBreadcrumb(data, "namespace", "Namespace"),
  },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: OrganizationNotFound,
  component: NamespaceDetailPage,
});

function NamespaceDetailPage() {
  const { organizationId, namespaceId } = Route.useParams();
  const data = Route.useLoaderData();

  if (isAccessDeniedRouteData(data)) {
    return <NamespaceAccessDenied />;
  }

  const { namespace, organization, permissions } = data;

  return (
    <div className="space-y-6">
      <PageHeader title={namespace.name} />

      <NamespaceDetailInfo
        namespace={namespace}
        permissions={permissions}
        organizationId={organizationId}
        organizationName={organization.name}
      />

      <NamespaceProjectsList
        projects={namespace.projects || []}
        isLoading={false}
        error={null}
        organizationId={organizationId}
        namespaceId={namespaceId}
        namespacePermissions={permissions}
      />

      <ScopedDocumentsList
        scope="namespace"
        documents={namespace.documents || []}
        isLoading={false}
        error={null}
      />

      <NamespaceDangerZone
        namespace={namespace}
        permissions={permissions}
        organizationId={organizationId}
      />
    </div>
  );
}
