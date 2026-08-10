import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";

import {
  OrganizationEditForm,
  OrganizationNotFound,
} from "@/components/organizations";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { PageHeader } from "@/components/shared/page-header";
import { v1OrganizationGetOptions } from "@/lib/api/query-options";
import { ResourceType } from "@/lib/auth/permissions";
import { requirePermissionBeforeLoad } from "@/lib/auth/require-auth";
import { namedEntityBreadcrumb } from "@/lib/breadcrumb";
import { loadOrganization } from "@/lib/route-data";
import { withRouteErrors } from "@/lib/route-errors";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationId/edit"
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
    breadcrumb: (data) => namedEntityBreadcrumb(data, "Organization"),
  },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: OrganizationNotFound,
  component: OrganizationEditPage,
});

function OrganizationEditPage() {
  const { organizationId } = Route.useParams();
  const { data: organization } = useSuspenseQuery(
    v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title="Edit Organization"
        description="Update the organization details below."
      />

      <OrganizationEditForm
        organization={organization}
        organizationId={organizationId}
      />
    </div>
  );
}
