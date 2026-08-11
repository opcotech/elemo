import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { ProjectActivityPage } from "@/components/projects";

const projectRoute = getRouteApi(
  "/_authenticated/namespaces/$namespaceId/projects/$projectId"
);

export const Route = createFileRoute(
  "/_authenticated/namespaces/$namespaceId/projects/$projectId/activity"
)({
  staticData: { breadcrumb: "Activity" },
  component: ProjectActivityRoute,
});

function ProjectActivityRoute() {
  const { namespace, project } = projectRoute.useLoaderData();
  return <ProjectActivityPage namespace={namespace} project={project} />;
}
