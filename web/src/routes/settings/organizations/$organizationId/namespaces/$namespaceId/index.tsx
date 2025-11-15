import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import {
  NamespaceDetailInfo,
  NamespaceDocumentsList,
  NamespaceProjectsList,
} from "@/components/namespaces";
import {
  OrganizationDetailError,
  OrganizationDetailSkeleton,
  OrganizationNotFound,
} from "@/components/organizations";
import { PageHeader } from "@/components/page-header";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import {
  isNotFound,
  v1NamespaceGetOptions,
  v1OrganizationGetOptions,
} from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

type RouteParams = {
  organizationId: string;
  namespaceId: string;
};

export const Route = createFileRoute(
  "/settings/organizations/$organizationId/namespaces/$namespaceId/"
)({
  beforeLoad: requireAuthBeforeLoad,
  component: NamespaceDetailPage,
});

function NamespaceDetailPage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const { organizationId, namespaceId } = Route.useParams() as RouteParams;

  const {
    data: namespace,
    isLoading: isLoadingNamespace,
    error: namespaceError,
  } = useQuery(
    v1NamespaceGetOptions({
      path: {
        id: namespaceId,
      },
    })
  );

  const {
    data: organization,
    isLoading: isLoadingOrg,
    error: orgError,
  } = useQuery(
    v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    })
  );

  const {
    data: namespacePermissions,
    isLoading: isNamespacePermissionsLoading,
  } = usePermissions(withResourceType(ResourceType.Namespace, namespaceId));

  const hasNamespaceReadPermission = can(namespacePermissions, "read");

  const isLoading = isLoadingNamespace || isLoadingOrg;
  const error = namespaceError || orgError;

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
    ]);
  }, [setBreadcrumbsFromItems, organization, namespace]);

  if (isLoading) {
    return <OrganizationDetailSkeleton />;
  }

  if (isNotFound(error) || !namespace || !organization) {
    return <OrganizationNotFound />;
  }

  if (error) {
    return <OrganizationDetailError />;
  }

  if (!isNamespacePermissionsLoading && !hasNamespaceReadPermission) {
    return (
      <div className="space-y-6">
        <PageHeader title="Access Denied" />
        <div className="text-muted-foreground">
          You do not have permission to view this namespace.
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader title={namespace.name} />

      <NamespaceDetailInfo
        namespace={namespace}
        organizationId={organizationId}
        organizationName={organization.name}
      />

      {!isNamespacePermissionsLoading && hasNamespaceReadPermission && (
        <>
          <NamespaceProjectsList
            projects={namespace.projects || []}
            isLoading={false}
            error={null}
          />

          <NamespaceDocumentsList
            documents={namespace.documents || []}
            isLoading={false}
            error={null}
          />
        </>
      )}
    </div>
  );
}
