import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { NamespaceDocumentsPage } from "@/components/namespaces/namespace-pages";
import { documentLibrarySearchSchema } from "@/lib/documents/library";

const namespaceRoute = getRouteApi("/_authenticated/namespaces/$namespaceId");

export const Route = createFileRoute(
  "/_authenticated/namespaces/$namespaceId/documents"
)({
  staticData: { breadcrumb: "Documents" },
  validateSearch: documentLibrarySearchSchema,
  component: NamespaceDocumentsRoute,
});

function NamespaceDocumentsRoute() {
  const { namespace } = namespaceRoute.useLoaderData();
  const search = Route.useSearch();
  return <NamespaceDocumentsPage namespace={namespace} search={search} />;
}
