import { queryOptions, useQuery, useQueryClient } from "@tanstack/react-query";
import type { QueryClient } from "@tanstack/react-query";

import { collectListedPage, cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1NamespacesGetOptions } from "@/lib/api/query-options";
import type { AccessibleNamespace as ApiAccessibleNamespace } from "@/lib/api/types";
import { cacheProfiles } from "@/lib/query-client";

export interface AccessibleNamespace extends ApiAccessibleNamespace {
  organizationId: string;
  organizationName: string;
}

export interface AccessibleWorkspace {
  organizations: ApiAccessibleNamespace["organization"][];
  namespaces: AccessibleNamespace[];
}

export const accessibleNamespacesQueryKey = [
  "elemo",
  "accessible-namespaces",
] as const;

function uniqueOrganizations(
  namespaces: AccessibleNamespace[]
): AccessibleWorkspace["organizations"] {
  const byId = new Map<string, AccessibleNamespace["organization"]>();
  for (const namespace of namespaces) {
    if (!byId.has(namespace.organization.id)) {
      byId.set(namespace.organization.id, namespace.organization);
    }
  }
  return [...byId.values()];
}

export function accessibleNamespacesOptions(queryClient: QueryClient) {
  return queryOptions<AccessibleWorkspace>({
    queryKey: accessibleNamespacesQueryKey,
    queryFn: async (): Promise<AccessibleWorkspace> => {
      const page = await collectListedPage(async (pageToken) =>
        queryClient.fetchQuery({
          ...v1NamespacesGetOptions({
            query: cursorPageQuery(pageToken),
          }),
          staleTime: 0,
        })
      );
      const namespaces = page.items.map((namespace) => ({
        ...namespace,
        organizationId: namespace.organization.id,
        organizationName: namespace.organization.name,
      }));

      return {
        organizations: uniqueOrganizations(namespaces),
        namespaces,
      };
    },
    ...cacheProfiles.reference,
  });
}

export function useAccessibleNamespaces() {
  const queryClient = useQueryClient();
  return useQuery(accessibleNamespacesOptions(queryClient));
}
