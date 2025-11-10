import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useMemo, useState } from "react";

import { NamespacesList } from "@/components/namespaces";
import {
  OrganizationDangerZone,
  OrganizationDetailError,
  OrganizationDetailInfo,
  OrganizationDetailSkeleton,
  OrganizationMembersList,
  OrganizationNotFound,
} from "@/components/organizations";
import { PageHeader } from "@/components/page-header";
import { RolesList } from "@/components/roles";
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
  v1OrganizationsNamespacesGetOptions,
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

  const {
    data: namespaces,
    isLoading: isLoadingNamespaces,
    error: namespacesError,
  } = useQuery(
    v1OrganizationsNamespacesGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

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

  const { data: orgPermissions, isLoading: isOrgPermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Organization, organizationId));

  const hasOrgReadPermission = can(orgPermissions, "read");

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
      <PageHeader title={organization.name} />

      <OrganizationDetailInfo organization={organization} />

      {!isOrgPermissionsLoading && hasOrgReadPermission && (
        <>
          <NamespacesList
            namespaces={namespaces || []}
            isLoading={isLoadingNamespaces}
            error={namespacesError}
            organizationId={organizationId}
          />

          <OrganizationMembersList
            members={processedMembers}
            isLoading={isLoadingMembers}
            error={membersError}
            currentUserId={currentUserId}
            organizationId={organizationId}
          />

          <RolesList
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
