import { createFileRoute } from "@tanstack/react-router";

import { RoleCreateForm } from "@/components/roles/role-create-form";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { PageHeader } from "@/components/ui/page-header";
import { Action } from "@/lib/auth/permissions";
import { withRouteErrors } from "@/lib/route-errors";
import { requireOrganizationAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/roles/new"
)({
  beforeLoad: ({ context, params }) =>
    requireOrganizationAction(
      context.queryClient,
      params.organizationSlug,
      Action.RoleManage
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      requireOrganizationAction(
        context.queryClient,
        params.organizationSlug,
        Action.RoleManage
      )
    ),
  staticData: { breadcrumb: "Create role" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: OrganizationRoleCreatePage,
});

function OrganizationRoleCreatePage() {
  const organization = Route.useLoaderData();

  return (
    <div className="space-y-6">
      <PageHeader
        title="Create Role"
        description="Create a new role for this organization."
      />

      <RoleCreateForm
        organizationId={organization.id}
        organizationSlug={organization.slug}
      />
    </div>
  );
}
