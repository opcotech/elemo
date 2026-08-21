import { createFileRoute } from "@tanstack/react-router";

import { NamespaceDangerZone } from "@/components/namespaces/namespace-danger-zone";
import { NamespaceDetailInfo } from "@/components/namespaces/namespace-detail-info";
import { NamespaceProjectsList } from "@/components/namespaces/namespace-projects-list";
import {
  SettingsAccessDenied,
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { PageHeader } from "@/components/ui/page-header";
import { entityBreadcrumb } from "@/lib/breadcrumb";
import { loadNamespaceDetail } from "@/lib/route-data";
import { isAccessDeniedRouteData, withRouteErrors } from "@/lib/route-errors";

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
  notFoundComponent: SettingsNotFound,
  component: NamespaceDetailPage,
});

function NamespaceDetailPage() {
  const { organizationId, namespaceId } = Route.useParams();
  const data = Route.useLoaderData();

  if (isAccessDeniedRouteData(data)) {
    return <SettingsAccessDenied resource="namespace" />;
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
        organizationId={organizationId}
        namespaceId={namespaceId}
        namespacePermissions={permissions}
      />

      <NamespaceDangerZone
        namespace={namespace}
        permissions={permissions}
        organizationId={organizationId}
      />
    </div>
  );
}
