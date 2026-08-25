import { createFileRoute } from "@tanstack/react-router";

import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { TeamEditForm } from "@/components/teams/team-edit-form";
import { PageHeader } from "@/components/ui/page-header";
import { Action } from "@/lib/auth/permissions";
import { loadOrganizationTeam } from "@/lib/route-data";
import { withRouteErrors } from "@/lib/route-errors";
import { requireOrganizationAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/teams/$teamId/edit"
)({
  beforeLoad: ({ context, params }) =>
    requireOrganizationAction(
      context.queryClient,
      params.organizationSlug,
      Action.TeamManage
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      loadOrganizationTeam(
        context.queryClient,
        params.organizationSlug,
        params.teamId
      )
    ),
  staticData: { breadcrumb: "Edit team" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: OrganizationTeamEditPage,
});

function OrganizationTeamEditPage() {
  const { teamId } = Route.useParams();
  const { organization, team } = Route.useLoaderData();

  return (
    <div className="space-y-6">
      <PageHeader
        title="Edit Team"
        description="Update the team details and members."
      />

      <TeamEditForm
        team={team}
        organizationId={organization.id}
        organizationSlug={organization.slug}
        teamId={teamId}
      />
    </div>
  );
}
