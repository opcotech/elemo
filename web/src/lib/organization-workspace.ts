import type { QueryClient } from "@tanstack/react-query";

import { cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1OrganizationGetOptions,
  v1OrganizationMembersGetOptions,
  v1OrganizationRolesGetOptions,
  v1OrganizationTeamsGetOptions,
  v1OrganizationsNamespacesGetOptions,
} from "@/lib/api/query-options";
import { organizationRefPath } from "@/lib/api/refs";
import type { Organization } from "@/lib/api/types";
import { Action, ResourceType, can } from "@/lib/auth/permissions";
import { loadResourcePermissions } from "@/lib/entity-context";

export async function loadOrganization(
  queryClient: QueryClient,
  organizationRef: string
): Promise<Organization> {
  const organization = await queryClient.fetchQuery(
    v1OrganizationGetOptions({ path: organizationRefPath(organizationRef) })
  );
  queryClient.setQueryData(
    v1OrganizationGetOptions({
      path: organizationRefPath(organization.id),
    }).queryKey,
    organization
  );
  return organization;
}

export async function loadOrganizationWorkspace(
  queryClient: QueryClient,
  organizationRef: string
) {
  const organization = await loadOrganization(queryClient, organizationRef);
  const permissions = await loadResourcePermissions(
    queryClient,
    ResourceType.Organization,
    organization.id
  );

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

  const orgPath = organizationRefPath(organization.id);
  const [membersPage, namespacesPage, rolesPage, teamsPage] = await Promise.all(
    [
      queryClient.fetchQuery(
        v1OrganizationMembersGetOptions({
          path: orgPath,
          query: cursorPageQuery(),
        })
      ),
      queryClient.fetchQuery(
        v1OrganizationsNamespacesGetOptions({
          path: orgPath,
          query: cursorPageQuery(),
        })
      ),
      queryClient.fetchQuery(
        v1OrganizationRolesGetOptions({
          path: orgPath,
          query: cursorPageQuery(),
        })
      ),
      queryClient.fetchQuery(
        v1OrganizationTeamsGetOptions({
          path: orgPath,
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
