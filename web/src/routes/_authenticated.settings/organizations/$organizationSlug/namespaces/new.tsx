import { createFileRoute } from "@tanstack/react-router";

import { NamespaceCreateForm } from "@/components/namespaces/namespace-create-form";
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
  "/_authenticated/settings/organizations/$organizationSlug/namespaces/new"
)({
  beforeLoad: ({ context, params }) =>
    requireOrganizationAction(
      context.queryClient,
      params.organizationSlug,
      Action.NamespaceCreate
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      requireOrganizationAction(
        context.queryClient,
        params.organizationSlug,
        Action.NamespaceCreate
      )
    ),
  staticData: {
    breadcrumb: "Create namespace",
  },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: OrganizationNamespaceCreatePage,
});

function OrganizationNamespaceCreatePage() {
  const organization = Route.useLoaderData();

  return (
    <div className="space-y-6">
      <PageHeader
        title="Create Namespace"
        description="Create a new namespace for this organization."
      />

      <NamespaceCreateForm
        organizationId={organization.id}
        organizationSlug={organization.slug}
      />
    </div>
  );
}
