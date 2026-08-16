import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";

import { NamespaceCreateForm } from "@/components/namespaces/namespace-create-form";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { PageHeader } from "@/components/ui/page-header";
import { v1OrganizationGetOptions } from "@/lib/api/query-options";
import { ResourceType } from "@/lib/auth/permissions";
import { requirePermissionBeforeLoad } from "@/lib/auth/require-auth";
import { loadOrganization } from "@/lib/route-data";
import { withRouteErrors } from "@/lib/route-errors";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationId/namespaces/new"
)({
  beforeLoad: ({ context, params }) =>
    requirePermissionBeforeLoad({
      queryClient: context.queryClient,
      resourceType: ResourceType.Organization,
      permissionKind: "write",
      resourceId: params.organizationId,
    }),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      loadOrganization(context.queryClient, params.organizationId)
    ),
  staticData: {
    breadcrumb: "Create namespace",
  },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: OrganizationNamespaceCreatePage,
});

function OrganizationNamespaceCreatePage() {
  const { organizationId } = Route.useParams();
  useSuspenseQuery(
    v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title="Create Namespace"
        description="Create a new namespace for this organization."
      />

      <NamespaceCreateForm organizationId={organizationId} />
    </div>
  );
}
