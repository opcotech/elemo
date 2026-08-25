import { Navigate, createFileRoute, getRouteApi } from "@tanstack/react-router";

const namespaceRoute = getRouteApi(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug"
);

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug/administration"
)({
  component: NamespaceAdministrationRedirect,
});

function NamespaceAdministrationRedirect() {
  const { namespace, organization } = namespaceRoute.useLoaderData();

  return (
    <Navigate
      to="/settings/organizations/$organizationSlug/namespaces/$namespaceSlug"
      params={{
        organizationSlug: organization.slug,
        namespaceSlug: namespace.slug,
      }}
      replace
    />
  );
}
