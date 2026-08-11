import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { ProjectOverviewPage } from "@/components/projects";

const projectRoute = getRouteApi(
  "/_authenticated/namespaces/$namespaceId/projects/$projectId"
);

export const Route = createFileRoute(
  "/_authenticated/namespaces/$namespaceId/projects/$projectId/"
)({
  component: ProjectOverviewRoute,
});

function ProjectOverviewRoute() {
  const { namespace, project } = projectRoute.useLoaderData();
  return <ProjectOverviewPage namespace={namespace} project={project} />;
}
