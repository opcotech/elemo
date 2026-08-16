import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";

import type { AccessibleNamespace } from "@/lib/api/accessible-namespaces";
import { useAccessibleNamespaces } from "@/lib/api/accessible-namespaces";
import { collectedListQuery, cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1OrganizationMembersGetOptions } from "@/lib/api/query-options";
import { v1OrganizationMembersGet } from "@/lib/api/sdk";

export function useOrganizationMembersForNamespace(
  namespaceId: string | undefined,
  options?: { enabled?: boolean }
) {
  const enabled = options?.enabled ?? true;
  const { data: accessibleWorkspace } = useAccessibleNamespaces();
  const organizationId = accessibleWorkspace?.namespaces.find(
    (namespace: AccessibleNamespace) => namespace.id === namespaceId
  )?.organizationId;

  const membersOptions = v1OrganizationMembersGetOptions({
    path: { id: organizationId ?? "" },
  });
  const { data: membersPage } = useQuery({
    ...collectedListQuery(membersOptions, async (pageToken, signal) => {
      const { data } = await v1OrganizationMembersGet({
        path: { id: organizationId ?? "" },
        query: cursorPageQuery(pageToken),
        signal,
        throwOnError: true,
      });
      return data;
    }),
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
