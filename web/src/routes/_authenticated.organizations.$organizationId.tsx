import { createFileRoute } from "@tanstack/react-router";

import {
  OrganizationDetailError,
  OrganizationDetailSkeleton,
  OrganizationNotFound,
  OrganizationOverviewPage,
} from "@/components/organizations";
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
  pendingComponent: OrganizationDetailSkeleton,
  errorComponent: OrganizationDetailError,
  notFoundComponent: OrganizationNotFound,
  component: OrganizationDetailRoute,
});

function OrganizationDetailRoute() {
  return <OrganizationOverviewPage data={Route.useLoaderData()} />;
}
