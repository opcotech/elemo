import { Outlet, createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/organizations")({
  staticData: {
    breadcrumb: { label: "Organizations", href: "/organizations" },
  },
  component: OrganizationsLayoutRoute,
});

function OrganizationsLayoutRoute() {
  return <Outlet />;
}
