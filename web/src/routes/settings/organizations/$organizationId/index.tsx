import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useMemo, useState } from "react";

import {
  OrganizationDangerZone,
  OrganizationDetailError,
  OrganizationDetailHeader,
  OrganizationDetailInfo,
  OrganizationDetailSkeleton,
  OrganizationMembersList,
  OrganizationNotFound,
  OrganizationRolesList,
} from "@/components/organizations";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import {
  isNotFound,
  v1OrganizationGetOptions,
  v1OrganizationMembersGetOptions,
  v1OrganizationRolesGetOptions,
} from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";
import { getUser } from "@/lib/auth/session";

export const Route = createFileRoute(
  "/settings/organizations/$organizationId/"
)({
  beforeLoad: requireAuthBeforeLoad,
  component: OrganizationDetailPage,
});

function OrganizationDetailPage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const { organizationId } = Route.useParams();
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);

  const {
    data: organization,
    isLoading,
    error,
  } = useQuery(
    v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

  // Fetch organization members
  const {
    data: members,
    isLoading: isLoadingMembers,
    error: membersError,
  } = useQuery(
    v1OrganizationMembersGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

  // Fetch organization roles
  const {
    data: roles,
    isLoading: isLoadingRoles,
    error: rolesError,
  } = useQuery(
    v1OrganizationRolesGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

  // Check if user has read permission (i.e., is a member of the organization)
  const { data: orgPermissions, isLoading: isOrgPermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Organization, organizationId));

  const hasOrgReadPermission = can(orgPermissions, "read");

  // Get current user ID
  useEffect(() => {
    const loadCurrentUser = async () => {
      const user = await getUser();
      if (user) {
        setCurrentUserId(user.id);
      }
    };
    loadCurrentUser();
  }, []);

  // No need to sort here - sorting is handled in OrganizationMembersList
  const processedMembers = useMemo(() => {
    return members || [];
  }, [members]);

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
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems, organization]);

  if (isLoading) {
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
      <OrganizationDetailHeader title={organization.name} />

      <OrganizationDetailInfo organization={organization} />

      {/* Only show members and roles if user has read permission (is a member) */}
      {!isOrgPermissionsLoading && hasOrgReadPermission && (
        <>
          <OrganizationMembersList
            members={processedMembers}
            isLoading={isLoadingMembers}
            error={membersError}
            currentUserId={currentUserId}
            organizationId={organizationId}
          />

          <OrganizationRolesList
            roles={roles || []}
            isLoading={isLoadingRoles}
            error={rolesError}
            organizationId={organizationId}
          />
        </>
      )}

      {organization.status === "active" && (
        <OrganizationDangerZone organization={organization} />
      )}
    </div>
  );
}
