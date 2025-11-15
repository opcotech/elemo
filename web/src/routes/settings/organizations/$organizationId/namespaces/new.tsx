import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import { NamespaceCreateForm } from "@/components/namespaces";
import {
  OrganizationDetailError,
  OrganizationDetailSkeleton,
  OrganizationNotFound,
} from "@/components/organizations";
import { PageHeader } from "@/components/page-header";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { ResourceType } from "@/hooks/use-permissions";
import { useRequirePermission } from "@/hooks/use-require-permission";
import { isNotFound, v1OrganizationGetOptions } from "@/lib/api";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

type RouteParams = {
  organizationId: string;
};

export const Route = createFileRoute(
  "/settings/organizations/$organizationId/namespaces/new"
)({
  beforeLoad: requireAuthBeforeLoad,
  component: OrganizationNamespaceCreatePage,
});

function OrganizationNamespaceCreatePage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const { organizationId } = Route.useParams() as RouteParams;

  // Check organization write permission
  const { isLoading: isCheckingOrgPermission } = useRequirePermission({
    resourceType: ResourceType.Organization,
    permissionKind: "write",
    resourceId: () => organizationId,
  });

  const {
    data: organization,
    isLoading,
    error,
  } = useQuery({
    ...v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    }),
    enabled: !isCheckingOrgPermission,
  });

  useEffect(() => {
    if (!organization) return;

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
        label: "Create Namespace",
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems, organization]);

  if (isCheckingOrgPermission || isLoading) {
    return <OrganizationDetailSkeleton />;
  }

  if (isNotFound(error) || !organization) {
    return <OrganizationNotFound />;
  }

  if (error) {
    return <OrganizationDetailError />;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Create Namespace"
        description="Create a new namespace for this organization."
      />

      <NamespaceCreateForm organizationId={organizationId} />
    </div>
  );
}
