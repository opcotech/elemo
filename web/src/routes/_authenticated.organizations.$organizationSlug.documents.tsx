import { createFileRoute } from "@tanstack/react-router";

import { OrganizationDetailError } from "@/components/organizations/organization-detail-error";
import { OrganizationNotFound } from "@/components/organizations/organization-not-found";
import { OrganizationDocumentsPage } from "@/components/organizations/organization-pages";
import { DetailPageSkeleton } from "@/components/ui/detail-card";
import { documentLibrarySearchSchema } from "@/lib/documents/library";
import { loadOrganizationWorkspace } from "@/lib/organization-workspace";
import { withRouteErrors } from "@/lib/route-errors";
import { requireOrganizationSlug } from "@/lib/route-identity";

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/documents"
)({
  staticData: { breadcrumb: "Documents" },
  validateSearch: documentLibrarySearchSchema,
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
  component: OrganizationDocumentsRoute,
});

function OrganizationDocumentsRoute() {
  const search = Route.useSearch();
  return (
    <OrganizationDocumentsPage data={Route.useLoaderData()} search={search} />
  );
}
