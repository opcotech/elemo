import type { QueryClient } from "@tanstack/react-query";

import { cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1OrganizationGetOptions,
  v1OrganizationMembersGetOptions,
  v1OrganizationRolesGetOptions,
  v1OrganizationTeamsGetOptions,
  v1OrganizationsNamespacesGetOptions,
} from "@/lib/api/query-options";
import { Action, ResourceType, can } from "@/lib/auth/permissions";
import { loadResourcePermissions } from "@/lib/entity-context";

export async function loadOrganization(
  queryClient: QueryClient,
  organizationId: string
) {
  // Prefer fetchQuery over ensureQueryData so invalidated (stale) loader
  // queries actually refetch after mutations. ensureQueryData returns any
  // cached value and never waits for a refresh.
  return queryClient.fetchQuery(
    v1OrganizationGetOptions({ path: { id: organizationId } })
  );
}

export async function loadOrganizationWorkspace(
  queryClient: QueryClient,
  organizationId: string
) {
  const [organization, permissions] = await Promise.all([
    loadOrganization(queryClient, organizationId),
    loadResourcePermissions(
      queryClient,
      ResourceType.Organization,
      organizationId
    ),
  ]);

  if (!can(permissions, Action.OrganizationRead)) {
    return {
      organization,
      permissions,
      members: [],
      namespaces: [],
      roles: [],
      teams: [],
      hasReadAccess: false as const,
    };
  }

  const [membersPage, namespacesPage, rolesPage, teamsPage] = await Promise.all(
    [
      queryClient.fetchQuery(
        v1OrganizationMembersGetOptions({
          path: { id: organizationId },
          query: cursorPageQuery(),
        })
      ),
      queryClient.fetchQuery(
        v1OrganizationsNamespacesGetOptions({
          path: { id: organizationId },
          query: cursorPageQuery(),
        })
      ),
      queryClient.fetchQuery(
        v1OrganizationRolesGetOptions({
          path: { id: organizationId },
          query: cursorPageQuery(),
        })
      ),
      queryClient.fetchQuery(
        v1OrganizationTeamsGetOptions({
          path: { id: organizationId },
          query: cursorPageQuery(),
        })
      ),
    ]
  );

  return {
    organization,
    permissions,
    members: membersPage.items ?? [],
    namespaces: namespacesPage.items ?? [],
    roles: rolesPage.items ?? [],
    teams: teamsPage.items ?? [],
    hasReadAccess: true as const,
  };
}

export type OrganizationWorkspaceData = Awaited<
  ReturnType<typeof loadOrganizationWorkspace>
>;
