import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";

import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { TeamEditForm } from "@/components/teams/team-edit-form";
import { PageHeader } from "@/components/ui/page-header";
import {
  v1OrganizationGetOptions,
  v1OrganizationTeamGetOptions,
} from "@/lib/api/query-options";
import { Action, ResourceType } from "@/lib/auth/permissions";
import { requirePermissionBeforeLoad } from "@/lib/auth/require-auth";
import { loadOrganizationTeam } from "@/lib/route-data";
import { withRouteErrors } from "@/lib/route-errors";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationId/teams/$teamId/edit"
)({
  beforeLoad: ({ context, params }) =>
    requirePermissionBeforeLoad({
      queryClient: context.queryClient,
      resourceType: ResourceType.Organization,
      action: Action.TeamManage,
      resourceId: params.organizationId,
    }),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      loadOrganizationTeam(
        context.queryClient,
        params.organizationId,
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
  const { organizationId, teamId } = Route.useParams();
  useSuspenseQuery(
    v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

  const { data: team } = useSuspenseQuery(
    v1OrganizationTeamGetOptions({
      path: {
        id: organizationId,
        team_id: teamId,
      },
    })
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title="Edit Team"
        description="Update the team details and members."
      />

      <TeamEditForm
        team={team}
        organizationId={organizationId}
        teamId={teamId}
      />
    </div>
  );
}
