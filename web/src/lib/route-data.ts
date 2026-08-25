import type { QueryClient } from "@tanstack/react-query";

import { accessibleNamespacesOptions } from "@/lib/api/accessible-namespaces";
import { cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1NamespaceGetOptions,
  v1NotificationsGetOptions,
  v1OrganizationRoleGetOptions,
  v1OrganizationTeamGetOptions,
  v1OrganizationsGetOptions,
} from "@/lib/api/query-options";
import { namespaceRefPath, organizationRefPath } from "@/lib/api/refs";
import { Action, ResourceType } from "@/lib/auth/permissions";
import {
  loadResourcePermissions,
  requireResourcePermission,
} from "@/lib/entity-context";
import { loadProjectByKey } from "@/lib/operational-route-data";
import {
  loadOrganization,
  loadOrganizationWorkspace,
} from "@/lib/organization-workspace";
import { withRouteErrors } from "@/lib/route-errors";

export { loadOrganization };

export const loadOrganizationDetail = loadOrganizationWorkspace;

export async function loadNamespaceHierarchy(
  queryClient: QueryClient,
  organizationRef: string,
  namespaceRef: string
) {
  return withRouteErrors(async () => {
    const [organization, namespace] = await Promise.all([
      loadOrganization(queryClient, organizationRef),
      queryClient.fetchQuery(
        v1NamespaceGetOptions({
          path: namespaceRefPath(organizationRef, namespaceRef),
        })
      ),
    ]);

    queryClient.setQueryData(
      v1NamespaceGetOptions({
        path: namespaceRefPath(namespace.organization.id, namespace.id),
      }).queryKey,
      namespace
    );

    return { organization, namespace };
  });
}

export async function loadNamespaceDetail(
  queryClient: QueryClient,
  organizationRef: string,
  namespaceRef: string
) {
  const hierarchy = await loadNamespaceHierarchy(
    queryClient,
    organizationRef,
    namespaceRef
  );
  const permissions = await requireResourcePermission(
    queryClient,
    ResourceType.Namespace,
    hierarchy.namespace.id,
    Action.NamespaceRead
  );
  return { ...hierarchy, permissions };
}

export async function loadProjectHierarchy(
  queryClient: QueryClient,
  organizationRef: string,
  namespaceRef: string,
  projectKey: string
) {
  return withRouteErrors(async () => {
    const { organization, namespace } = await loadNamespaceHierarchy(
      queryClient,
      organizationRef,
      namespaceRef
    );

    const project = await loadProjectByKey(
      queryClient,
      organizationRef,
      namespaceRef,
      projectKey
    );

    return { organization, namespace, project };
  });
}

export async function loadProjectDetail(
  queryClient: QueryClient,
  organizationRef: string,
  namespaceRef: string,
  projectKey: string
) {
  return withRouteErrors(async () => {
    const { organization, namespace } = await loadNamespaceHierarchy(
      queryClient,
      organizationRef,
      namespaceRef
    );

    const project = await loadProjectByKey(
      queryClient,
      organizationRef,
      namespaceRef,
      projectKey
    );

    const permissions = await requireResourcePermission(
      queryClient,
      ResourceType.Project,
      project.id,
      Action.ProjectRead
    );

    return { organization, namespace, project, permissions };
  });
}

export async function loadOrganizationRole(
  queryClient: QueryClient,
  organizationRef: string,
  roleId: string
) {
  const [organization, role] = await Promise.all([
    loadOrganization(queryClient, organizationRef),
    queryClient.fetchQuery(
      v1OrganizationRoleGetOptions({
        path: { ...organizationRefPath(organizationRef), role_id: roleId },
      })
    ),
  ]);
  return { organization, role };
}

export async function loadOrganizationTeam(
  queryClient: QueryClient,
  organizationRef: string,
  teamId: string
) {
  const [organization, team] = await Promise.all([
    loadOrganization(queryClient, organizationRef),
    queryClient.fetchQuery(
      v1OrganizationTeamGetOptions({
        path: { ...organizationRefPath(organizationRef), team_id: teamId },
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
