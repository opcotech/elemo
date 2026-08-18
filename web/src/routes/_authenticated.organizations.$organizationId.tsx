import { Outlet, createFileRoute } from "@tanstack/react-router";

import { OrganizationDetailError } from "@/components/organizations/organization-detail-error";
import { OrganizationNotFound } from "@/components/organizations/organization-not-found";
import { DetailPageSkeleton } from "@/components/ui/detail-card";
import { entityBreadcrumb } from "@/lib/breadcrumb";
import { loadOrganizationWorkspace } from "@/lib/organization-workspace";
import { withRouteErrors } from "@/lib/route-errors";

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationId"
)({
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      loadOrganizationWorkspace(context.queryClient, params.organizationId)
    ),
  staticData: {
    breadcrumb: (data) =>
      entityBreadcrumb(data, "organization", "Organization"),
  },
  pendingComponent: DetailPageSkeleton,
  errorComponent: OrganizationDetailError,
  notFoundComponent: OrganizationNotFound,
  component: OrganizationLayoutRoute,
});

function OrganizationLayoutRoute() {
  return <Outlet />;
}
