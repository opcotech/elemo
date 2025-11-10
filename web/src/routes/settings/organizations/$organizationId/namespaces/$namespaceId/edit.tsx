import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import { NamespaceEditForm } from "@/components/namespaces";
import {
  OrganizationDetailError,
  OrganizationDetailSkeleton,
  OrganizationNotFound,
} from "@/components/organizations";
import { PageHeader } from "@/components/page-header";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { ResourceType } from "@/hooks/use-permissions";
import { useRequirePermission } from "@/hooks/use-require-permission";
import {
  isNotFound,
  v1NamespaceGetOptions,
  v1OrganizationGetOptions,
} from "@/lib/api";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

type RouteParams = {
  organizationId: string;
  namespaceId: string;
};

export const Route = createFileRoute(
  "/settings/organizations/$organizationId/namespaces/$namespaceId/edit"
)({
  beforeLoad: requireAuthBeforeLoad,
  component: OrganizationNamespaceEditPage,
});

function OrganizationNamespaceEditPage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const { organizationId, namespaceId } = Route.useParams() as RouteParams;

  const { isLoading: isCheckingOrgPermission } = useRequirePermission({
    resourceType: ResourceType.Organization,
    permissionKind: "write",
    resourceId: () => organizationId,
  });

  const { isLoading: isCheckingNamespacePermission } = useRequirePermission({
    resourceType: ResourceType.Namespace,
    permissionKind: "write",
    resourceId: () => namespaceId,
  });

  const isCheckingPermission =
    isCheckingOrgPermission || isCheckingNamespacePermission;

  const {
    data: organization,
    isLoading: isLoadingOrg,
    error: orgError,
  } = useQuery({
    ...v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    }),
    enabled: !isCheckingPermission,
  });

  const {
    data: namespace,
    isLoading: isLoadingNamespace,
    error: namespaceError,
  } = useQuery({
    ...v1NamespaceGetOptions({
      path: {
        id: namespaceId,
      },
    }),
    enabled: !isCheckingPermission,
  });

  const isLoading = isLoadingOrg || isLoadingNamespace;
  const error = orgError || namespaceError;

  useEffect(() => {
    if (!organization || !namespace) return;

    setBreadcrumbsFromItems([
      {
        label: "Settings",
        href: "/settings",
        isNavigatable: true,
      },
      {
        label: "Organizations",
        href: "/settings/organizations",
        isNavigatable: true,
      },
      {
        label: organization.name,
        href: `/settings/organizations/${organization.id}`,
        isNavigatable: true,
      },
      {
        label: namespace.name,
        isNavigatable: false,
      },
      {
        label: "Edit",
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems, organization, namespace]);

  if (isCheckingPermission || isLoading) {
    return <OrganizationDetailSkeleton />;
  }

  if (isNotFound(error) || !organization || !namespace) {
    return <OrganizationNotFound />;
  }

  if (error) {
    return <OrganizationDetailError />;
  }

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
