import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import {
  OrganizationDetailError,
  OrganizationDetailHeader,
  OrganizationDetailSkeleton,
  OrganizationNotFound,
  OrganizationRoleCreateForm,
} from "@/components/organizations";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { ResourceType } from "@/hooks/use-permissions";
import { useRequirePermission } from "@/hooks/use-require-permission";
import { isNotFound, v1OrganizationGetOptions } from "@/lib/api";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

type RouteParams = {
  organizationId: string;
};

export const Route = createFileRoute(
  "/settings/organizations/$organizationId/roles/new"
)({
  beforeLoad: requireAuthBeforeLoad,
  component: OrganizationRoleCreatePage,
});

function OrganizationRoleCreatePage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const { organizationId } = Route.useParams() as RouteParams;

  // Check organization write permission
  const { isLoading: isCheckingOrgPermission } = useRequirePermission({
    resourceType: ResourceType.Organization,
    permissionKind: "write",
    resourceId: () => organizationId,
  });

  // Check role create permission
  const { isLoading: isCheckingRolePermission } = useRequirePermission({
    resourceType: ResourceType.Role,
    permissionKind: "create",
  });

  const isCheckingPermission =
    isCheckingOrgPermission || isCheckingRolePermission;

  // Fetch organization data for breadcrumbs
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
    // Don't fetch organization data until permission is confirmed
    enabled: !isCheckingPermission,
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
        label: "Create Role",
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems, organization]);

  // Show loading while checking permissions or loading organization
  if (isCheckingPermission || isLoading) {
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
      <OrganizationDetailHeader
        title="Create Role"
        description="Create a new role for this organization."
      />

      <OrganizationRoleCreateForm organizationId={organizationId} />
    </div>
  );
}
