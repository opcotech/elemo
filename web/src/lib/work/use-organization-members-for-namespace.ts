import { useQueries, useQuery } from "@tanstack/react-query";
import { useMemo } from "react";

import type { AccessibleNamespace } from "@/lib/api/accessible-namespaces";
import { useAccessibleNamespaces } from "@/lib/api/accessible-namespaces";
import { collectedListQuery, cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1OrganizationMembersGetOptions } from "@/lib/api/query-options";
import { v1OrganizationMembersGet } from "@/lib/api/sdk";
import type { OrganizationMember } from "@/lib/api/types";

function organizationMembersCollectedQuery(organizationId: string) {
  const membersOptions = v1OrganizationMembersGetOptions({
    path: { organizationRef: organizationId },
  });
  return collectedListQuery(membersOptions, async (pageToken, signal) => {
    const { data } = await v1OrganizationMembersGet({
      path: { organizationRef: organizationId },
      query: cursorPageQuery(pageToken),
      signal,
      throwOnError: true,
    });
    return data;
  });
}

export function useOrganizationMembersForNamespace(
  namespaceId: string | undefined,
  options?: { enabled?: boolean }
) {
  const enabled = options?.enabled ?? true;
  const { data: accessibleWorkspace } = useAccessibleNamespaces();
  const organizationId = accessibleWorkspace?.namespaces.find(
    (namespace: AccessibleNamespace) => namespace.id === namespaceId
  )?.organizationId;

  const { data: membersPage } = useQuery({
    ...organizationMembersCollectedQuery(organizationId ?? ""),
    enabled: Boolean(organizationId) && enabled,
  });

  return useMemo(
    () => ({
      organizationId,
      members: membersPage?.items ?? [],
    }),
    [membersPage, organizationId]
  );
}

/** Members from every organization the current user can access. */
export function useAccessibleOrganizationMembers(options?: {
  enabled?: boolean;
}) {
  const enabled = options?.enabled ?? true;
  const { data: accessibleWorkspace } = useAccessibleNamespaces();
  const organizationIds = useMemo(
    () =>
      (accessibleWorkspace?.organizations ?? []).map(
        (organization) => organization.id
      ),
    [accessibleWorkspace]
  );

  const memberQueries = useQueries({
    queries: organizationIds.map((organizationId) => ({
      ...organizationMembersCollectedQuery(organizationId),
      enabled,
    })),
  });

  const members = useMemo(() => {
    const byId = new Map<string, OrganizationMember>();
    for (const query of memberQueries) {
      for (const member of query.data?.items ?? []) {
        byId.set(member.id, member);
      }
    }
    return [...byId.values()];
  }, [memberQueries]);

  return { members };
}
