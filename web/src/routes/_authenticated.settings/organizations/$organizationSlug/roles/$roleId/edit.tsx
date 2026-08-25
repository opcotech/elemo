import { createFileRoute } from "@tanstack/react-router";

import { RoleEditFormWithPermissions } from "@/components/roles/role-edit-form";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { PageHeader } from "@/components/ui/page-header";
import { Action } from "@/lib/auth/permissions";
import { loadOrganizationRole } from "@/lib/route-data";
import { withRouteErrors } from "@/lib/route-errors";
import { requireOrganizationAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/roles/$roleId/edit"
)({
  beforeLoad: ({ context, params }) =>
    requireOrganizationAction(
      context.queryClient,
      params.organizationSlug,
      Action.RoleManage
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      loadOrganizationRole(
        context.queryClient,
        params.organizationSlug,
        params.roleId
      )
    ),
  staticData: { breadcrumb: "Edit role" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: OrganizationRoleEditPage,
});

function OrganizationRoleEditPage() {
  const { roleId } = Route.useParams();
  const { organization, role } = Route.useLoaderData();

  return (
    <div className="space-y-6">
      <PageHeader
        title="Edit Role"
        description="Update the role details below."
      />

      <RoleEditFormWithPermissions
        role={role}
        organizationId={organization.id}
        organizationSlug={organization.slug}
        roleId={roleId}
      />
    </div>
  );
}
