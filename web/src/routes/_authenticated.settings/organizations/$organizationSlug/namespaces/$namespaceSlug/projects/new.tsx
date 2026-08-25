import { createFileRoute } from "@tanstack/react-router";

import { ProjectCreateForm } from "@/components/projects/project-create-form";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { PageHeader } from "@/components/ui/page-header";
import { Action } from "@/lib/auth/permissions";
import { withRouteErrors } from "@/lib/route-errors";
import { requireNamespaceAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/new"
)({
  beforeLoad: ({ context, params }) =>
    requireNamespaceAction(
      context.queryClient,
      params.organizationSlug,
      params.namespaceSlug,
      Action.ProjectCreate
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      requireNamespaceAction(
        context.queryClient,
        params.organizationSlug,
        params.namespaceSlug,
        Action.ProjectCreate
      )
    ),
  staticData: { breadcrumb: "Create project" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: OrganizationNamespaceProjectCreatePage,
});

function OrganizationNamespaceProjectCreatePage() {
  const { organization, namespace } = Route.useLoaderData();

  return (
    <div className="space-y-6">
      <PageHeader
        title="Create Project"
        description={`Create a new project in ${namespace.name}.`}
      />

      <ProjectCreateForm
        organizationId={organization.id}
        organizationSlug={organization.slug}
        namespaceId={namespace.id}
        namespaceSlug={namespace.slug}
      />
    </div>
  );
}
