import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import {
  OrganizationDetailError,
  OrganizationDetailSkeleton,
  OrganizationNotFound,
  OrganizationRoleEditFormWithPermissions,
} from "@/components/organizations";
import { PageHeader } from "@/components/page-header";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { ResourceType } from "@/hooks/use-permissions";
import { useRequirePermission } from "@/hooks/use-require-permission";
import {
  isNotFound,
  v1OrganizationGetOptions,
  v1OrganizationRoleGetOptions,
} from "@/lib/api";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

type RouteParams = {
  organizationId: string;
  roleId: string;
};

export const Route = createFileRoute(
  "/settings/organizations/$organizationId/roles/$roleId/edit"
)({
  beforeLoad: requireAuthBeforeLoad,
  component: OrganizationRoleEditPage,
});

function OrganizationRoleEditPage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const { organizationId, roleId } = Route.useParams() as RouteParams;

  const { isLoading: isCheckingOrgPermission } = useRequirePermission({
    resourceType: ResourceType.Organization,
    permissionKind: "write",
    resourceId: () => organizationId,
  });

  const { isLoading: isCheckingRolePermission } = useRequirePermission({
    resourceType: ResourceType.Role,
    permissionKind: "write",
    resourceId: () => roleId,
  });

  const isCheckingPermission =
    isCheckingOrgPermission || isCheckingRolePermission;

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
    data: role,
    isLoading: isLoadingRole,
    error: roleError,
  } = useQuery({
    ...v1OrganizationRoleGetOptions({
      path: {
        id: organizationId,
        role_id: roleId,
      },
    }),
    enabled: !isCheckingPermission,
  });

  const isLoading = isLoadingOrg || isLoadingRole;
  const error = orgError || roleError;

  useEffect(() => {
    if (!organization || !role) return;

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
        label: role.name,
        isNavigatable: false,
      },
      {
        label: "Edit",
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems, organization, role]);

  if (isCheckingPermission || isLoading) {
    return <OrganizationDetailSkeleton />;
  }

  if (isNotFound(error) || !organization || !role) {
    return <OrganizationNotFound />;
  }

  if (error) {
    return <OrganizationDetailError />;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Edit Role"
        description="Update the role details below."
      />

      <OrganizationRoleEditFormWithPermissions
        role={role}
        organizationId={organizationId}
        roleId={roleId}
      />
    </div>
  );
}
