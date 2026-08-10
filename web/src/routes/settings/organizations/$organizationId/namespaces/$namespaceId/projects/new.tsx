import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import {
  OrganizationDetailError,
  OrganizationDetailSkeleton,
  OrganizationNotFound,
} from "@/components/organizations";
import { PageHeader } from "@/components/page-header";
import { ProjectCreateForm } from "@/components/projects";
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
  "/settings/organizations/$organizationId/namespaces/$namespaceId/projects/new"
)({
  beforeLoad: requireAuthBeforeLoad,
  component: OrganizationNamespaceProjectCreatePage,
});

function OrganizationNamespaceProjectCreatePage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const { organizationId, namespaceId } = Route.useParams() as RouteParams;

  const { isLoading: isCheckingNamespacePermission } = useRequirePermission({
    resourceType: ResourceType.Namespace,
    permissionKind: "write",
    resourceId: () => namespaceId,
  });

  const {
    data: organization,
    isLoading: isLoadingOrganization,
    error: organizationError,
  } = useQuery({
    ...v1OrganizationGetOptions({
      path: {
        id: organizationId,
      },
    }),
    enabled: !isCheckingNamespacePermission,
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
    enabled: !isCheckingNamespacePermission,
  });

  const isLoading = isLoadingOrganization || isLoadingNamespace;
  const error = organizationError || namespaceError;

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
        href: `/settings/organizations/${organization.id}/namespaces/${namespace.id}`,
        isNavigatable: true,
      },
      {
        label: "Create Project",
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems, organization, namespace]);

  if (isCheckingNamespacePermission || isLoading) {
    return <OrganizationDetailSkeleton />;
  }

  if (isNotFound(error)) {
    return <OrganizationNotFound />;
  }

  if (error) {
    return <OrganizationDetailError />;
  }

  if (!organization || !namespace) {
    return <OrganizationNotFound />;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Create Project"
        description={`Create a new project in ${namespace.name}.`}
      />

      <ProjectCreateForm
        organizationId={organizationId}
        namespaceId={namespaceId}
      />
    </div>
  );
}
