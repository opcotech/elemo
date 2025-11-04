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
  isNotFound,
  v1OrganizationGetOptions,
  v1OrganizationMembersGetOptions,
  v1OrganizationRolesGetOptions,
} from "@/lib/api";
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

  const processedMembers = useMemo(() => {
    if (!members) return [];

    return [...members].sort((a, b) => {
      if (a.status !== b.status) {
        if (a.status === "deleted") return 1;
        if (b.status === "deleted") return -1;
      }

      const aName = `${a.first_name} ${a.last_name}`.toLowerCase();
      const bName = `${b.first_name} ${b.last_name}`.toLowerCase();
      return aName.localeCompare(bName);
    });
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

      <OrganizationMembersList
        members={processedMembers}
        isLoading={isLoadingMembers}
        error={membersError}
        currentUserId={currentUserId}
      />

      <OrganizationRolesList
        roles={roles || []}
        isLoading={isLoadingRoles}
        error={rolesError}
        organizationId={organizationId}
      />

      {organization.status === "active" && (
        <OrganizationDangerZone organization={organization} />
      )}
    </div>
  );
}
