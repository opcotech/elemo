import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { ProjectActivityPage } from "@/components/projects/project-pages";

const projectRoute = getRouteApi(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey"
);

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey/activity"
)({
  staticData: { breadcrumb: "Activity" },
  component: ProjectActivityRoute,
});

function ProjectActivityRoute() {
  const { namespace, project } = projectRoute.useLoaderData();
  return <ProjectActivityPage namespace={namespace} project={project} />;
}
