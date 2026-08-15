import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { NamespaceOverviewPage } from "@/components/namespaces/namespace-pages";

const namespaceRoute = getRouteApi("/_authenticated/namespaces/$namespaceId");

export const Route = createFileRoute(
  "/_authenticated/namespaces/$namespaceId/"
)({
  component: NamespaceOverviewRoute,
});

function NamespaceOverviewRoute() {
  const { namespace, organization } = namespaceRoute.useLoaderData();
  return (
    <NamespaceOverviewPage namespace={namespace} organization={organization} />
  );
}
