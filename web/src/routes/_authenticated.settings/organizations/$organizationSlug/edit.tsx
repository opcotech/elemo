import { createFileRoute } from "@tanstack/react-router";

import { OrganizationEditForm } from "@/components/organizations/organization-edit-form";
import { OrganizationNotFound } from "@/components/organizations/organization-not-found";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { PageHeader } from "@/components/ui/page-header";
import { Action } from "@/lib/auth/permissions";
import { namedEntityBreadcrumb } from "@/lib/breadcrumb";
import { withRouteErrors } from "@/lib/route-errors";
import { requireOrganizationAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/edit"
)({
  beforeLoad: ({ context, params }) =>
    requireOrganizationAction(
      context.queryClient,
      params.organizationSlug,
      Action.OrganizationUpdate
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      requireOrganizationAction(
        context.queryClient,
        params.organizationSlug,
        Action.OrganizationUpdate
      )
    ),
  staticData: {
    breadcrumb: (data) => namedEntityBreadcrumb(data, "Organization"),
  },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: OrganizationNotFound,
  component: OrganizationEditPage,
});

function OrganizationEditPage() {
  const organization = Route.useLoaderData();

  return (
    <div className="space-y-6">
      <PageHeader
        title="Edit Organization"
        description="Update the organization details below."
      />

      <OrganizationEditForm
        organization={organization}
        organizationId={organization.id}
      />
    </div>
  );
}
