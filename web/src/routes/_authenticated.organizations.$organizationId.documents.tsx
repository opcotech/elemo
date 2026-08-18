import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { OrganizationDocumentsPage } from "@/components/organizations/organization-pages";
import { documentLibrarySearchSchema } from "@/lib/documents/library";

const organizationRoute = getRouteApi(
  "/_authenticated/organizations/$organizationId"
);

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationId/documents"
)({
  staticData: { breadcrumb: "Documents" },
  validateSearch: documentLibrarySearchSchema,
  component: OrganizationDocumentsRoute,
});

function OrganizationDocumentsRoute() {
  const search = Route.useSearch();
  return (
    <OrganizationDocumentsPage
      data={organizationRoute.useLoaderData()}
      search={search}
    />
  );
}
