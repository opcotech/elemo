import { createFileRoute } from "@tanstack/react-router";

import { OrganizationDetailError } from "@/components/organizations/organization-detail-error";
import { OrganizationNotFound } from "@/components/organizations/organization-not-found";
import { OrganizationOverviewPage } from "@/components/organizations/organization-pages";
import { DetailPageSkeleton } from "@/components/ui/detail-card";
import { loadOrganizationWorkspace } from "@/lib/organization-workspace";
import { withRouteErrors } from "@/lib/route-errors";
import { requireOrganizationSlug } from "@/lib/route-identity";

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/"
)({
  loader: ({ context, params }) =>
    withRouteErrors(async () => {
      requireOrganizationSlug(params.organizationSlug);
      return loadOrganizationWorkspace(
        context.queryClient,
        params.organizationSlug
      );
    }),
  pendingComponent: DetailPageSkeleton,
  errorComponent: OrganizationDetailError,
  notFoundComponent: OrganizationNotFound,
  component: OrganizationOverviewRoute,
});

function OrganizationOverviewRoute() {
  return <OrganizationOverviewPage data={Route.useLoaderData()} />;
}
