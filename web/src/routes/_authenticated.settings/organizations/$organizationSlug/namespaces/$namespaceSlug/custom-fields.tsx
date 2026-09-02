import { createFileRoute } from "@tanstack/react-router";

import { CustomFieldDefinitionManager } from "@/components/custom-fields/definition-manager";
import {
  SettingsEntityDetailError,
  SettingsEntityDetailSkeleton,
} from "@/components/settings/settings-entity-detail-state";
import { SettingsNotFound } from "@/components/settings/settings-not-found";
import { PageHeader } from "@/components/ui/page-header";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import { Action, can } from "@/lib/auth/permissions";
import { withRouteErrors } from "@/lib/route-errors";
import { requireNamespaceAction } from "@/lib/route-permissions";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/custom-fields"
)({
  beforeLoad: ({ context, params }) =>
    requireNamespaceAction(
      context.queryClient,
      params.organizationSlug,
      params.namespaceSlug,
      Action.NamespaceRead
    ),
  loader: ({ context, params }) =>
    withRouteErrors(() =>
      requireNamespaceAction(
        context.queryClient,
        params.organizationSlug,
        params.namespaceSlug,
        Action.NamespaceRead
      )
    ),
  staticData: { breadcrumb: "Custom fields" },
  pendingComponent: SettingsEntityDetailSkeleton,
  errorComponent: SettingsEntityDetailError,
  notFoundComponent: SettingsNotFound,
  component: NamespaceCustomFieldsPage,
});

function NamespaceCustomFieldsPage() {
  const { namespace } = Route.useLoaderData();
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Namespace, namespace.id)
  );
  const canManage = can(permissions, Action.CustomFieldManage);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Custom fields"
        description="Define typed Issue fields for this namespace. Inherited organization fields are visible but not editable here."
      />

      <CustomFieldDefinitionManager
        scopeId={namespace.id}
        scopeType="Namespace"
        canManage={canManage}
      />
    </div>
  );
}
