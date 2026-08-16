import type { QueryClient } from "@tanstack/react-query";

import { collectListedPage, cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1OrganizationGetOptions,
  v1OrganizationMembersGetOptions,
  v1OrganizationRolesGetOptions,
  v1OrganizationsNamespacesGetOptions,
} from "@/lib/api/query-options";
import { ResourceType, can } from "@/lib/auth/permissions";
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

  if (!can(permissions, "read")) {
    return {
      organization,
      permissions,
      members: [],
      namespaces: [],
      roles: [],
      hasReadAccess: false as const,
    };
  }

  const [membersPage, namespaces, roles] = await Promise.all([
    collectListedPage(async (pageToken) =>
      queryClient.fetchQuery(
        v1OrganizationMembersGetOptions({
          path: { id: organizationId },
          query: cursorPageQuery(pageToken),
        })
      )
    ),
    collectListedPage(async (pageToken) =>
      queryClient.fetchQuery(
        v1OrganizationsNamespacesGetOptions({
          path: { id: organizationId },
          query: cursorPageQuery(pageToken),
        })
      )
    ),
    collectListedPage(async (pageToken) =>
      queryClient.fetchQuery(
        v1OrganizationRolesGetOptions({
          path: { id: organizationId },
          query: cursorPageQuery(pageToken),
        })
      )
    ),
  ]);

  return {
    organization,
    permissions,
    members: membersPage.items,
    namespaces: namespaces.items,
    roles: roles.items,
    hasReadAccess: true as const,
  };
}

export type OrganizationWorkspaceData = Awaited<
  ReturnType<typeof loadOrganizationWorkspace>
>;
