import type { QueryClient } from "@tanstack/react-query";
import { notFound } from "@tanstack/react-router";

import { accessibleNamespacesOptions } from "@/lib/api/accessible-namespaces";
import {
  v1NamespaceGetOptions,
  v1NotificationsGetOptions,
  v1OrganizationRoleGetOptions,
  v1OrganizationsGetOptions,
  v1ProjectGetOptions,
} from "@/lib/api/query-options";
import { ResourceType } from "@/lib/auth/permissions";
import {
  loadResourcePermissions,
  requireResourcePermission,
} from "@/lib/entity-context";
import {
  loadOrganization,
  loadOrganizationWorkspace,
} from "@/lib/organization-workspace";

export { loadOrganization };

export const loadOrganizationDetail = loadOrganizationWorkspace;

export async function loadNamespaceHierarchy(
  queryClient: QueryClient,
  organizationId: string,
  namespaceId: string
) {
  const [organization, namespace] = await Promise.all([
    loadOrganization(queryClient, organizationId),
    queryClient.fetchQuery(
      v1NamespaceGetOptions({ path: { id: namespaceId } })
    ),
  ]);

  if (!organization.namespaces.includes(namespaceId)) {
    throw notFound();
  }

  return { organization, namespace };
}

export async function loadNamespaceDetail(
  queryClient: QueryClient,
  organizationId: string,
  namespaceId: string
) {
  const hierarchy = await loadNamespaceHierarchy(
    queryClient,
    organizationId,
    namespaceId
  );
  const permissions = await requireResourcePermission(
    queryClient,
    ResourceType.Namespace,
    namespaceId
  );
  return { ...hierarchy, permissions };
}

export async function loadProjectHierarchy(
  queryClient: QueryClient,
  organizationId: string,
  namespaceId: string,
  projectId: string
) {
  const { organization, namespace } = await loadNamespaceHierarchy(
    queryClient,
    organizationId,
    namespaceId
  );

  if (!namespace.projects.some((project) => project.id === projectId)) {
    throw notFound();
  }

  const project = await queryClient.fetchQuery(
    v1ProjectGetOptions({ path: { id: projectId } })
  );

  return { organization, namespace, project };
}

export async function loadProjectDetail(
  queryClient: QueryClient,
  organizationId: string,
  namespaceId: string,
  projectId: string
) {
  const { organization } = await loadNamespaceHierarchy(
    queryClient,
    organizationId,
    namespaceId
  );

  // Namespace membership can change immediately after project create; refetch
  // when stale/invalidated so the detail route does not 404 on a stale list.
  const namespace = await queryClient.fetchQuery(
    v1NamespaceGetOptions({ path: { id: namespaceId } })
  );

  if (!namespace.projects.some((project) => project.id === projectId)) {
    throw notFound();
  }

  const permissions = await requireResourcePermission(
    queryClient,
    ResourceType.Project,
    projectId
  );

  const project = await queryClient.fetchQuery(
    v1ProjectGetOptions({ path: { id: projectId } })
  );

  return { organization, namespace, project, permissions };
}

export async function loadOrganizationRole(
  queryClient: QueryClient,
  organizationId: string,
  roleId: string
) {
  const [organization, role] = await Promise.all([
    loadOrganization(queryClient, organizationId),
    queryClient.fetchQuery(
      v1OrganizationRoleGetOptions({
        path: { id: organizationId, role_id: roleId },
      })
    ),
  ]);
  return { organization, role };
}

export async function loadAllNamespaces(queryClient: QueryClient) {
  const [workspace] = await Promise.all([
    queryClient.fetchQuery(accessibleNamespacesOptions(queryClient)),
    loadOrganizationsWithPermissions(queryClient),
  ]);

  return workspace;
}

export async function loadOrganizations(
  queryClient: QueryClient,
  includeRowPermissions = false
) {
  const organizations = await queryClient.fetchQuery(
    v1OrganizationsGetOptions()
  );
  const permissionLoads = [
    loadResourcePermissions(queryClient, ResourceType.Organization),
  ];

  if (includeRowPermissions) {
    permissionLoads.push(
      ...organizations.map((organization) =>
        loadResourcePermissions(
          queryClient,
          ResourceType.Organization,
          organization.id
        )
      )
    );
  }

  await Promise.all(permissionLoads);
  return organizations;
}

export async function loadOrganizationsWithPermissions(
  queryClient: QueryClient
) {
  return loadOrganizations(queryClient, true);
}

export function prefetchAuthenticatedChrome(queryClient: QueryClient) {
  void Promise.all([
    queryClient.prefetchQuery(v1OrganizationsGetOptions()),
    queryClient.prefetchQuery(v1NotificationsGetOptions()),
  ]);
}
