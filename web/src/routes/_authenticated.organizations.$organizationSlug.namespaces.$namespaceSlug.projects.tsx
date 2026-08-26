import { Outlet, createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug/projects"
)({
  staticData: { breadcrumb: "Projects" },
  component: NamespaceProjectsLayoutRoute,
});

function NamespaceProjectsLayoutRoute() {
  return <Outlet />;
}
