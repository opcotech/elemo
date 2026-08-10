import { Outlet, createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/namespaces/$namespaceId/projects"
)({
  staticData: { breadcrumb: "Projects" },
  component: NamespaceProjectsLayoutRoute,
});

function NamespaceProjectsLayoutRoute() {
  return <Outlet />;
}
