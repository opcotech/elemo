import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { NamespaceOverviewPage } from "@/components/namespaces/namespace-pages";

const namespaceRoute = getRouteApi(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug"
);

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug/"
)({
  component: NamespaceOverviewRoute,
});

function NamespaceOverviewRoute() {
  const { namespace, organization } = namespaceRoute.useLoaderData();
  return (
    <NamespaceOverviewPage namespace={namespace} organization={organization} />
  );
}
