import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { NamespaceDocumentsPage } from "@/components/namespaces/namespace-pages";
import { documentLibrarySearchSchema } from "@/lib/documents/library";

const namespaceRoute = getRouteApi(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug"
);

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug/documents"
)({
  staticData: { breadcrumb: "Documents" },
  validateSearch: documentLibrarySearchSchema,
  component: NamespaceDocumentsRoute,
});

function NamespaceDocumentsRoute() {
  const { namespace, organization } = namespaceRoute.useLoaderData();
  const search = Route.useSearch();
  return (
    <NamespaceDocumentsPage
      namespace={namespace}
      organization={organization}
      search={search}
    />
  );
}
