import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { NamespaceProjectsPage } from "@/components/namespaces/namespace-pages";

const namespaceRoute = getRouteApi(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug"
);

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/"
)({
  component: NamespaceProjectsRoute,
});

function NamespaceProjectsRoute() {
  const { namespace, organization } = namespaceRoute.useLoaderData();
  return (
    <NamespaceProjectsPage namespace={namespace} organization={organization} />
  );
}
