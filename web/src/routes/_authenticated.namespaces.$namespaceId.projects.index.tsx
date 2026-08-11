import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { NamespaceProjectsPage } from "@/components/namespaces";

const namespaceRoute = getRouteApi("/_authenticated/namespaces/$namespaceId");

export const Route = createFileRoute(
  "/_authenticated/namespaces/$namespaceId/projects/"
)({
  component: NamespaceProjectsRoute,
});

function NamespaceProjectsRoute() {
  const { namespace, organization } = namespaceRoute.useLoaderData();
  return (
    <NamespaceProjectsPage namespace={namespace} organization={organization} />
  );
}
