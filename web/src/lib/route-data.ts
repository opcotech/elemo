import type { QueryClient } from "@tanstack/react-query";
import { notFound } from "@tanstack/react-router";

import { accessibleNamespacesOptions } from "@/lib/api/accessible-namespaces";
import { collectListedPage, cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1NamespaceGetOptions,
  v1NamespacesProjectsGetOptions,
  v1NotificationsGetOptions,
  v1OrganizationRoleGetOptions,
  v1OrganizationTeamGetOptions,
  v1OrganizationsGetOptions,
  v1OrganizationsNamespacesGetOptions,
  v1ProjectGetOptions,
} from "@/lib/api/query-options";
import { Action, ResourceType } from "@/lib/auth/permissions";
import {
  loadResourcePermissions,
  requireResourcePermission,
} from "@/lib/entity-context";
import {
  loadOrganization,
  loadOrganizationWorkspace,
} from "@/lib/organization-workspace";

export { loadOrganization };

async function requireProjectInNamespace(
  queryClient: QueryClient,
  namespaceId: string,
  projectId: string
) {
  const projectsPage = await collectListedPage(async (pageToken) =>
    queryClient.fetchQuery(
      v1NamespacesProjectsGetOptions({
        path: { id: namespaceId },
        query: cursorPageQuery(pageToken),
      })
    )
  );

  if (!projectsPage.items.some((item) => item.id === projectId)) {
    throw notFound();
  }
}

export const loadOrganizationDetail = loadOrganizationWorkspace;

export async function loadNamespaceHierarchy(
  queryClient: QueryClient,
  organizationId: string,
  namespaceId: string
) {
  const [organization, namespace, namespacesPage] = await Promise.all([
    loadOrganization(queryClient, organizationId),
    queryClient.fetchQuery(
      v1NamespaceGetOptions({ path: { id: namespaceId } })
    ),
    collectListedPage(async (pageToken) =>
      queryClient.fetchQuery(
        v1OrganizationsNamespacesGetOptions({
          path: { id: organizationId },
          query: cursorPageQuery(pageToken),
        })
      )
    ),
  ]);

  if (!namespacesPage.items.some((item) => item.id === namespaceId)) {
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
    namespaceId,
    Action.NamespaceRead
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

  await requireProjectInNamespace(queryClient, namespaceId, projectId);

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

  // Permission-filtered project lists omit rows the actor cannot read. Check
  // project.read first so a direct URL becomes Access Denied instead of a 404.
  const permissions = await requireResourcePermission(
    queryClient,
    ResourceType.Project,
    projectId,
    Action.ProjectRead
  );

  await requireProjectInNamespace(queryClient, namespaceId, projectId);

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

export async function loadOrganizationTeam(
  queryClient: QueryClient,
  organizationId: string,
  teamId: string
) {
  const [organization, team] = await Promise.all([
    loadOrganization(queryClient, organizationId),
    queryClient.fetchQuery(
      v1OrganizationTeamGetOptions({
        path: { id: organizationId, team_id: teamId },
      })
    ),
  ]);
  return { organization, team };
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
  const organizationsPage = await queryClient.fetchQuery(
    v1OrganizationsGetOptions({
      query: cursorPageQuery(),
    })
  );
  const organizations = organizationsPage.items;
  const permissionLoads = [
    loadResourcePermissions(queryClient, ResourceType.Installation),
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
    queryClient.prefetchQuery(
      v1OrganizationsGetOptions({
        query: cursorPageQuery(),
      })
    ),
    queryClient.prefetchQuery(
      v1NotificationsGetOptions({
        query: cursorPageQuery(),
      })
    ),
  ]);
}
