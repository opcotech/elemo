import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { OrganizationOverviewPage } from "@/components/organizations/organization-pages";

const organizationRoute = getRouteApi(
  "/_authenticated/organizations/$organizationId"
);

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationId/"
)({
  component: OrganizationOverviewRoute,
});

function OrganizationOverviewRoute() {
  return <OrganizationOverviewPage data={organizationRoute.useLoaderData()} />;
}
