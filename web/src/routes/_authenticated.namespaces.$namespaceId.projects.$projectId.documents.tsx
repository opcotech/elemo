import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { ProjectDocumentsPage } from "@/components/projects/project-pages";

const projectRoute = getRouteApi(
  "/_authenticated/namespaces/$namespaceId/projects/$projectId"
);

export const Route = createFileRoute(
  "/_authenticated/namespaces/$namespaceId/projects/$projectId/documents"
)({
  staticData: { breadcrumb: "Documents" },
  component: ProjectDocumentsRoute,
});

function ProjectDocumentsRoute() {
  const { namespace, project } = projectRoute.useLoaderData();
  return <ProjectDocumentsPage namespace={namespace} project={project} />;
}
