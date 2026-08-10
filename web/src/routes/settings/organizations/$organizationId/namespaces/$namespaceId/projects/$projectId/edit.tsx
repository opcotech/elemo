import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import {
  OrganizationDetailError,
  OrganizationDetailSkeleton,
  OrganizationNotFound,
} from "@/components/organizations";
import { PageHeader } from "@/components/page-header";
import { ProjectEditForm } from "@/components/projects";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { ResourceType } from "@/hooks/use-permissions";
import { useRequirePermission } from "@/hooks/use-require-permission";
import {
  isNotFound,
  v1NamespaceGetOptions,
  v1OrganizationGetOptions,
  v1ProjectGetOptions,
} from "@/lib/api";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

type RouteParams = {
  organizationId: string;
  namespaceId: string;
  projectId: string;
};

export const Route = createFileRoute(
  "/settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId/edit"
)({
  beforeLoad: requireAuthBeforeLoad,
  component: ProjectEditPage,
});

function ProjectEditPage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const { organizationId, namespaceId, projectId } =
    Route.useParams() as RouteParams;

  const { isLoading: isCheckingPermission } = useRequirePermission({
    resourceType: ResourceType.Project,
    permissionKind: "write",
    resourceId: () => projectId,
  });

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

  const {
    data: project,
    isLoading: isLoadingProject,
    error: projectError,
  } = useQuery({
    ...v1ProjectGetOptions({
      path: {
        id: projectId,
      },
    }),
    enabled: !isCheckingPermission,
  });

  const isLoading = isLoadingOrg || isLoadingNamespace || isLoadingProject;
  const error = orgError || namespaceError || projectError;

  useEffect(() => {
    if (!organization || !namespace || !project) return;

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
        label: project.name,
        href: `/settings/organizations/${organization.id}/namespaces/${namespace.id}/projects/${project.id}`,
        isNavigatable: true,
      },
      {
        label: "Edit",
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems, organization, namespace, project]);

  if (isCheckingPermission || isLoading) {
    return <OrganizationDetailSkeleton />;
  }

  if (isNotFound(error) || !organization || !namespace || !project) {
    return <OrganizationNotFound />;
  }

  if (error) {
    return <OrganizationDetailError />;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Edit Project"
        description="Update the project details below."
      />

      <ProjectEditForm
        project={project}
        organizationId={organizationId}
        namespaceId={namespaceId}
      />
    </div>
  );
}
