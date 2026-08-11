import { Outlet, createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/namespaces")({
  staticData: {
    breadcrumb: { label: "Namespaces", href: "/namespaces" },
  },
  component: NamespacesLayoutRoute,
});

function NamespacesLayoutRoute() {
  return <Outlet />;
}
