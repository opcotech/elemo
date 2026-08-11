import { Outlet, createFileRoute } from "@tanstack/react-router";

import { entityBreadcrumb } from "@/lib/breadcrumb";
import { loadNamespaceOperationalContext } from "@/lib/operational-route-data";

export const Route = createFileRoute("/_authenticated/namespaces/$namespaceId")(
  {
    staticData: {
      breadcrumb: (data) => entityBreadcrumb(data, "namespace", "Namespace"),
    },
    loader: ({ context, params }) =>
      loadNamespaceOperationalContext(context.queryClient, params.namespaceId),
    component: NamespaceLayoutRoute,
  }
);

function NamespaceLayoutRoute() {
  return <Outlet />;
}
