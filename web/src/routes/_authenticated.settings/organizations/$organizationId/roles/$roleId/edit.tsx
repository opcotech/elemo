import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";

import { OrganizationNotFound } from "@/components/organizations";
import { RoleEditFormWithPermissions } from "@/components/roles";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { PageHeader } from "@/components/shared/page-header";
import {
  v1OrganizationGetOptions,
  v1OrganizationRoleGetOptions,
} from "@/lib/api/query-options";
import { ResourceType } from "@/lib/auth/permissions";
import { requirePermissionBeforeLoad } from "@/lib/auth/require-auth";
import { loadOrganizationRole } from "@/lib/route-data";
import { withRouteErrors } from "@/lib/route-errors";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationId/roles/$roleId/edit"
)({
  beforeLoad: ({ context, params }) =>
    Promise.all([
      requirePermissionBeforeLoad({
        queryClient: context.queryClient,
        resourceType: ResourceType.Organization,
        permissionKind: "write",
        resourceId: params.organizationId,
      }),
      requirePermissionBeforeLoad({
        queryClient: context.queryClient,
        resourceType: ResourceType.Role,
        permissionKind: "write",
        resourceId: params.roleId,
      }),
    ]),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      loadOrganizationRole(
        context.queryClient,
        params.organizationId,
        params.roleId
      )
    ),
  staticData: { breadcrumb: "Edit role" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: OrganizationNotFound,
  component: OrganizationRoleEditPage,
});

function OrganizationRoleEditPage() {
  const { organizationId, roleId } = Route.useParams();
  useSuspenseQuery(
    v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

  const { data: role } = useSuspenseQuery(
    v1OrganizationRoleGetOptions({
      path: {
        id: organizationId,
        role_id: roleId,
      },
    })
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title="Edit Role"
        description="Update the role details below."
      />

      <RoleEditFormWithPermissions
        role={role}
        organizationId={organizationId}
        roleId={roleId}
      />
    </div>
  );
}
