import { createFileRoute } from "@tanstack/react-router";

import { CustomFieldDefinitionManager } from "@/components/custom-fields/definition-manager";
import { OrganizationNotFound } from "@/components/organizations/organization-not-found";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { PageHeader } from "@/components/ui/page-header";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import { Action, can } from "@/lib/auth/permissions";
import { withRouteErrors } from "@/lib/route-errors";
import { requireOrganizationAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/custom-fields"
)({
  beforeLoad: ({ context, params }) =>
    requireOrganizationAction(
      context.queryClient,
      params.organizationSlug,
      Action.OrganizationRead
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      requireOrganizationAction(
        context.queryClient,
        params.organizationSlug,
        Action.OrganizationRead
      )
    ),
  staticData: { breadcrumb: "Custom fields" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: OrganizationNotFound,
  component: OrganizationCustomFieldsPage,
});

function OrganizationCustomFieldsPage() {
  const organization = Route.useLoaderData();
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Organization, organization.id)
  );
  const canManage = can(permissions, Action.CustomFieldManage);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Custom fields"
        description="Define typed Issue fields for this organization. They apply to issues in every namespace and project."
      />

      <CustomFieldDefinitionManager
        scopeId={organization.id}
        scopeType="Organization"
        canManage={canManage}
      />
    </div>
  );
}
