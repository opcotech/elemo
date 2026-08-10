import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { WorkSurface } from "@/components/work";
import { workRouteSearchSchema } from "@/lib/work-route-search";

const projectRoute = getRouteApi(
  "/_authenticated/namespaces/$namespaceId/projects/$projectId"
);

export const Route = createFileRoute(
  "/_authenticated/namespaces/$namespaceId/projects/$projectId/work"
)({
  staticData: { breadcrumb: "Work" },
  validateSearch: workRouteSearchSchema,
  component: ProjectWorkRoute,
});

function ProjectWorkRoute() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const { namespace, project } = projectRoute.useLoaderData();
  return (
    <WorkSurface
      title={`${project.name} / Work`}
      description="Project-scoped work using the shared projection query."
      context={{ namespace: namespace.name, project: project.name }}
      scope={{
        type: "project",
        namespaceId: namespace.id,
        projectId: project.id,
      }}
      search={search}
      onSearchChange={(patch) =>
        void navigate({
          search: (previous) => ({ ...previous, ...patch }),
          replace: true,
        })
      }
    />
  );
}
