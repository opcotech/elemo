import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";

import { NamespaceEditForm } from "@/components/namespaces/namespace-edit-form";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { PageHeader } from "@/components/ui/page-header";
import {
  v1NamespaceGetOptions,
  v1OrganizationGetOptions,
} from "@/lib/api/query-options";
import { Action, ResourceType } from "@/lib/auth/permissions";
import { requirePermissionBeforeLoad } from "@/lib/auth/require-auth";
import { loadNamespaceHierarchy } from "@/lib/route-data";
import { withRouteErrors } from "@/lib/route-errors";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationId/namespaces/$namespaceId/edit"
)({
  beforeLoad: ({ context, params }) =>
    Promise.all([
      requirePermissionBeforeLoad({
        queryClient: context.queryClient,
        resourceType: ResourceType.Organization,
        action: Action.OrganizationRead,
        resourceId: params.organizationId,
      }),
      requirePermissionBeforeLoad({
        queryClient: context.queryClient,
        resourceType: ResourceType.Namespace,
        action: Action.NamespaceUpdate,
        resourceId: params.namespaceId,
      }),
    ]),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      loadNamespaceHierarchy(
        context.queryClient,
        params.organizationId,
        params.namespaceId
      )
    ),
  staticData: { breadcrumb: "Edit namespace" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: OrganizationNamespaceEditPage,
});

function OrganizationNamespaceEditPage() {
  const { organizationId, namespaceId } = Route.useParams();
  useSuspenseQuery(
    v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

  const { data: namespace } = useSuspenseQuery(
    v1NamespaceGetOptions({
      path: {
        id: namespaceId,
      },
    })
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title="Edit Namespace"
        description="Update the namespace details below."
      />

      <NamespaceEditForm
        namespace={namespace}
        organizationId={organizationId}
      />
    </div>
  );
}
