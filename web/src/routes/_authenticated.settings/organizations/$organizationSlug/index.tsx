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
import { zOrganizationStatus } from "@/lib/api/schemas";
import { entityBreadcrumb } from "@/lib/breadcrumb";
import { loadOrganizationWorkspace } from "@/lib/organization-workspace";
import { withRouteErrors } from "@/lib/route-errors";
import { requireOrganizationSlug } from "@/lib/route-identity";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/"
)({
  loader: ({ context, params }) =>
    withRouteErrors(async () => {
      requireOrganizationSlug(params.organizationSlug);
      return loadOrganizationWorkspace(
        context.queryClient,
        params.organizationSlug
      );
    }),
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
  const currentUserId = user?.id ?? null;
  const { organization, permissions, hasReadAccess } = Route.useLoaderData();
  const organizationId = organization.id;

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
            organizationId={organizationId}
            organizationSlug={organization.slug}
            organizationPermissions={permissions}
          />

          <OrganizationMembersList
            currentUserId={currentUserId}
            organizationId={organizationId}
            organizationPermissions={permissions}
          />

          <RolesList
            organizationId={organizationId}
            organizationSlug={organization.slug}
            organizationPermissions={permissions}
          />

          <TeamsList
            organizationId={organizationId}
            organizationSlug={organization.slug}
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
