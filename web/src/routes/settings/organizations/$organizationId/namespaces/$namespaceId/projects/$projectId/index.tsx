import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import {
  OrganizationDetailError,
  OrganizationDetailSkeleton,
  OrganizationNotFound,
} from "@/components/organizations";
import { PageHeader } from "@/components/page-header";
import {
  ProjectDangerZone,
  ProjectDetailInfo,
  ProjectDocumentsList,
  ProjectIssuesList,
} from "@/components/projects";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import {
  isNotFound,
  isPermissionDenied,
  v1NamespaceGetOptions,
  v1OrganizationGetOptions,
  v1ProjectGetOptions,
} from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

type RouteParams = {
  organizationId: string;
  namespaceId: string;
  projectId: string;
};

export const Route = createFileRoute(
  "/settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId/"
)({
  beforeLoad: requireAuthBeforeLoad,
  component: ProjectDetailPage,
});

function ProjectDetailPage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const { organizationId, namespaceId, projectId } =
    Route.useParams() as RouteParams;

  const {
    data: project,
    isLoading: isLoadingProject,
    error: projectError,
  } = useQuery(
    v1ProjectGetOptions({
      path: {
        id: projectId,
      },
    })
  );

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

  const { data: projectPermissions, isLoading: isProjectPermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Project, projectId));

  const hasProjectReadPermission = can(projectPermissions, "read");

  const isLoading = isLoadingProject || isLoadingNamespace || isLoadingOrg;
  const error = projectError || namespaceError || orgError;

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
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems, organization, namespace, project]);

  const accessDenied = (
    <div className="space-y-6">
      <PageHeader title="Access Denied" />
      <div className="text-muted-foreground">
        You do not have permission to view this project.
      </div>
    </div>
  );

  if (isLoading || isProjectPermissionsLoading) {
    return <OrganizationDetailSkeleton />;
  }

  if (isPermissionDenied(projectError)) {
    return accessDenied;
  }

  if (isNotFound(error) || !namespace || !organization) {
    return <OrganizationNotFound />;
  }

  if (!hasProjectReadPermission) {
    return accessDenied;
  }

  if (!project) {
    return <OrganizationNotFound />;
  }

  if (error) {
    return <OrganizationDetailError />;
  }

  return (
    <div className="space-y-6">
      <PageHeader title={project.name} />

      <ProjectDetailInfo
        project={project}
        organizationId={organizationId}
        namespaceId={namespaceId}
        namespaceName={namespace.name}
      />

      <ProjectDocumentsList
        documents={project.documents || []}
        isLoading={false}
        error={null}
      />

      <ProjectIssuesList
        issues={project.issues || []}
        isLoading={false}
        error={null}
      />

      <ProjectDangerZone
        project={project}
        organizationId={organizationId}
        namespaceId={namespaceId}
      />
    </div>
  );
}
