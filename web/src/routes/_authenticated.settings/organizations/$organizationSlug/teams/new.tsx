import { createFileRoute } from "@tanstack/react-router";

import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { TeamCreateForm } from "@/components/teams/team-create-form";
import { PageHeader } from "@/components/ui/page-header";
import { Action } from "@/lib/auth/permissions";
import { withRouteErrors } from "@/lib/route-errors";
import { requireOrganizationAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/teams/new"
)({
  beforeLoad: ({ context, params }) =>
    requireOrganizationAction(
      context.queryClient,
      params.organizationSlug,
      Action.TeamManage
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      requireOrganizationAction(
        context.queryClient,
        params.organizationSlug,
        Action.TeamManage
      )
    ),
  staticData: { breadcrumb: "Create team" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: OrganizationTeamCreatePage,
});

function OrganizationTeamCreatePage() {
  const organization = Route.useLoaderData();

  return (
    <div className="space-y-6">
      <PageHeader
        title="Create Team"
        description="Create a team that can hold grants as a principal."
      />

      <TeamCreateForm
        organizationId={organization.id}
        organizationSlug={organization.slug}
      />
    </div>
  );
}
