import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { NamespaceDocumentsPage } from "@/components/namespaces";

const namespaceRoute = getRouteApi("/_authenticated/namespaces/$namespaceId");

export const Route = createFileRoute(
  "/_authenticated/namespaces/$namespaceId/documents"
)({
  staticData: { breadcrumb: "Documents" },
  component: NamespaceDocumentsRoute,
});

function NamespaceDocumentsRoute() {
  const { namespace } = namespaceRoute.useLoaderData();
  return <NamespaceDocumentsPage namespace={namespace} />;
}
