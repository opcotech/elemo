import { Outlet, createFileRoute } from "@tanstack/react-router";

import { entityBreadcrumb } from "@/lib/breadcrumb";
import { loadProjectOperationalContext } from "@/lib/operational-route-data";

export const Route = createFileRoute(
  "/_authenticated/namespaces/$namespaceId/projects/$projectId"
)({
  staticData: {
    breadcrumb: (data) => entityBreadcrumb(data, "project", "Project"),
  },
  loader: ({ context, params }) =>
    loadProjectOperationalContext(
      context.queryClient,
      params.namespaceId,
      params.projectId
    ),
  component: ProjectLayoutRoute,
});

function ProjectLayoutRoute() {
  return <Outlet />;
}
