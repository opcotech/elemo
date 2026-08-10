import { createFileRoute } from "@tanstack/react-router";

import { NamespacesList } from "@/components/namespaces";
import {
  OrganizationDangerZone,
  OrganizationDetailInfo,
  OrganizationMembersList,
  OrganizationNotFound,
} from "@/components/organizations";
import { RolesList } from "@/components/roles";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { PageHeader } from "@/components/shared/page-header";
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
