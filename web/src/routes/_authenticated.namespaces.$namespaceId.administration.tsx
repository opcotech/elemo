import { Navigate, createFileRoute, getRouteApi } from "@tanstack/react-router";

const namespaceRoute = getRouteApi("/_authenticated/namespaces/$namespaceId");

export const Route = createFileRoute(
  "/_authenticated/namespaces/$namespaceId/administration"
)({
  component: NamespaceAdministrationRedirect,
});

function NamespaceAdministrationRedirect() {
  const { namespace, organization } = namespaceRoute.useLoaderData();

  return (
    <Navigate
      to="/settings/organizations/$organizationId/namespaces/$namespaceId"
      params={{
        organizationId: organization.id,
        namespaceId: namespace.id,
      }}
      replace
    />
  );
}
