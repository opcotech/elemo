import { Outlet, createFileRoute } from "@tanstack/react-router";

import { NotFound } from "@/components/shared/not-found";
import { entityBreadcrumb } from "@/lib/breadcrumb";
import { loadNamespaceOperationalContext } from "@/lib/operational-route-data";

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug"
)({
  staticData: {
    breadcrumb: (data) => entityBreadcrumb(data, "namespace", "Namespace"),
  },
  loader: ({ context, params }) =>
    loadNamespaceOperationalContext(
      context.queryClient,
      params.organizationSlug,
      params.namespaceSlug
    ),
  notFoundComponent: NotFound,
  component: NamespaceLayoutRoute,
});

function NamespaceLayoutRoute() {
  return <Outlet />;
}
