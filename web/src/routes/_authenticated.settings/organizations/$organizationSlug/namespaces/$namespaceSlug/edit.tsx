import { createFileRoute } from "@tanstack/react-router";

import { NamespaceEditForm } from "@/components/namespaces/namespace-edit-form";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { PageHeader } from "@/components/ui/page-header";
import { Action } from "@/lib/auth/permissions";
import { withRouteErrors } from "@/lib/route-errors";
import { requireNamespaceAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/edit"
)({
  beforeLoad: ({ context, params }) =>
    requireNamespaceAction(
      context.queryClient,
      params.organizationSlug,
      params.namespaceSlug,
      Action.NamespaceUpdate
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      requireNamespaceAction(
        context.queryClient,
        params.organizationSlug,
        params.namespaceSlug,
        Action.NamespaceUpdate
      )
    ),
  staticData: { breadcrumb: "Edit namespace" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: OrganizationNamespaceEditPage,
});

function OrganizationNamespaceEditPage() {
  const { organization, namespace } = Route.useLoaderData();

  return (
    <div className="space-y-6">
      <PageHeader
        title="Edit Namespace"
        description="Update the namespace details below."
      />

      <NamespaceEditForm
        namespace={namespace}
        organizationId={organization.id}
        organizationSlug={organization.slug}
      />
    </div>
  );
}
