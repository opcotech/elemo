import { Outlet, createFileRoute } from "@tanstack/react-router";

import { NotFound } from "@/components/shared/not-found";
import { entityBreadcrumb } from "@/lib/breadcrumb";
import { loadProjectOperationalContext } from "@/lib/operational-route-data";

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey"
)({
  staticData: {
    breadcrumb: (data) => entityBreadcrumb(data, "project", "Project"),
  },
  loader: ({ context, params }) =>
    loadProjectOperationalContext(
      context.queryClient,
      params.organizationSlug,
      params.namespaceSlug,
      params.projectKey
    ),
  notFoundComponent: NotFound,
  component: ProjectLayoutRoute,
});

function ProjectLayoutRoute() {
  return <Outlet />;
}
