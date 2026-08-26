import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { WorkSurface } from "@/components/work/work-surface";
import { workRouteSearchSchema } from "@/lib/work-route-search";

const projectRoute = getRouteApi(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey"
);

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey/work"
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
      description="Project-scoped work from the live issue API."
      context={{ namespace: namespace.name, project: project.name }}
      scope={{
        type: "project",
        namespaceId: namespace.id,
        organizationId: namespace.organization.id,
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
