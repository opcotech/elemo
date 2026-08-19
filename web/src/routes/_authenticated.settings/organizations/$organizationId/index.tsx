import { createFileRoute } from "@tanstack/react-router";

import { GrantCreateForm } from "@/components/grants/grant-create-form";
import { NamespacesList } from "@/components/namespaces/namespaces-list";
import { OrganizationDangerZone } from "@/components/organizations/organization-danger-zone";
import { OrganizationDetailInfo } from "@/components/organizations/organization-detail-info";
import { OrganizationMembersList } from "@/components/organizations/organization-members-list";
import { OrganizationNotFound } from "@/components/organizations/organization-not-found";
import { RolesList } from "@/components/roles/roles-list";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { TeamsList } from "@/components/teams/teams-list";
import { PageHeader } from "@/components/ui/page-header";
import { useAuth } from "@/hooks/use-auth";
import { entityBreadcrumb } from "@/lib/breadcrumb";
import { zOrganizationStatus } from "@/lib/client/zod.gen";
import { loadOrganizationWorkspace } from "@/lib/organization-workspace";
import { withRouteErrors } from "@/lib/route-errors";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationId/"
)({
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      loadOrganizationWorkspace(context.queryClient, params.organizationId)
    ),
  staticData: {
    breadcrumb: (data) =>
      entityBreadcrumb(data, "organization", "Organization"),
  },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: OrganizationNotFound,
  component: OrganizationDetailPage,
});

function OrganizationDetailPage() {
  const { user } = useAuth();
  const { organizationId } = Route.useParams();
  const currentUserId = user?.id ?? null;
  const {
    organization,
    members,
    namespaces,
    roles,
    teams,
    permissions,
    hasReadAccess,
  } = Route.useLoaderData();

  return (
    <div className="space-y-6">
      <PageHeader title={organization.name} />

      <OrganizationDetailInfo
        organization={organization}
        permissions={permissions}
      />

      {hasReadAccess && (
        <>
          <NamespacesList
            namespaces={namespaces}
            isLoading={false}
            error={null}
            organizationId={organizationId}
            organizationPermissions={permissions}
          />

          <OrganizationMembersList
            members={members}
            isLoading={false}
            error={null}
            currentUserId={currentUserId}
            organizationId={organizationId}
            organizationPermissions={permissions}
          />

          <RolesList
            roles={roles}
            isLoading={false}
            error={null}
            organizationId={organizationId}
            organizationPermissions={permissions}
          />

          <TeamsList
            teams={teams}
            isLoading={false}
            error={null}
            organizationId={organizationId}
            organizationPermissions={permissions}
          />

          <GrantCreateForm organizationId={organizationId} />
        </>
      )}

      {organization.status === zOrganizationStatus.enum.active && (
        <OrganizationDangerZone
          organization={organization}
          permissions={permissions}
        />
      )}
    </div>
  );
}
