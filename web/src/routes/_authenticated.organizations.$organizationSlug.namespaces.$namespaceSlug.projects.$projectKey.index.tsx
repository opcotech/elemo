import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { ProjectOverviewPage } from "@/components/projects/project-pages";

const projectRoute = getRouteApi(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey"
);

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey/"
)({
  component: ProjectOverviewRoute,
});

function ProjectOverviewRoute() {
  const { namespace, project } = projectRoute.useLoaderData();
  return <ProjectOverviewPage namespace={namespace} project={project} />;
}
