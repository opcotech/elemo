import { createFileRoute } from "@tanstack/react-router";

import { ProjectEditForm } from "@/components/projects/project-edit-form";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { PageHeader } from "@/components/ui/page-header";
import { Action } from "@/lib/auth/permissions";
import { withRouteErrors } from "@/lib/route-errors";
import { requireProjectAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey/edit"
)({
  beforeLoad: ({ context, params }) =>
    requireProjectAction(
      context.queryClient,
      params.organizationSlug,
      params.namespaceSlug,
      params.projectKey,
      Action.ProjectUpdate
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      requireProjectAction(
        context.queryClient,
        params.organizationSlug,
        params.namespaceSlug,
        params.projectKey,
        Action.ProjectUpdate
      )
    ),
  staticData: { breadcrumb: "Edit project" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: ProjectEditPage,
});

function ProjectEditPage() {
  const { organization, namespace, project } = Route.useLoaderData();

  return (
    <div className="space-y-6">
      <PageHeader
        title="Edit Project"
        description="Update the project details below."
      />

      <ProjectEditForm
        project={project}
        organizationId={organization.id}
        organizationSlug={organization.slug}
        namespaceId={namespace.id}
        namespaceSlug={namespace.slug}
      />
    </div>
  );
}
