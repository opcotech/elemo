import { queryOptions, useQuery, useQueryClient } from "@tanstack/react-query";
import type { QueryClient } from "@tanstack/react-query";

import { collectListedPage, cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1OrganizationsGetOptions,
  v1OrganizationsNamespacesGetOptions,
} from "@/lib/api/query-options";
import type { Namespace, Organization } from "@/lib/api/types";
import { cacheProfiles } from "@/lib/query-client";

export interface AccessibleNamespace extends Namespace {
  organization: Organization;
  organizationId: string;
  organizationName: string;
}

export interface AccessibleWorkspace {
  organizations: Organization[];
  namespaces: AccessibleNamespace[];
}

export const accessibleNamespacesQueryKey = [
  "elemo",
  "accessible-namespaces",
] as const;

export function accessibleNamespacesOptions(queryClient: QueryClient) {
  return queryOptions<AccessibleWorkspace>({
    queryKey: accessibleNamespacesQueryKey,
    queryFn: async (): Promise<AccessibleWorkspace> => {
      const organizationsPage = await collectListedPage(async (pageToken) =>
        queryClient.fetchQuery({
          ...v1OrganizationsGetOptions({
            query: cursorPageQuery(pageToken),
          }),
          staleTime: 0,
        })
      );
      const organizations = organizationsPage.items;
      const namespacesByOrganization = await Promise.all(
        organizations.map((organization) =>
          collectListedPage(async (pageToken) =>
            queryClient.fetchQuery({
              ...v1OrganizationsNamespacesGetOptions({
                path: { id: organization.id },
                query: cursorPageQuery(pageToken),
              }),
              staleTime: 0,
            })
          )
        )
      );

      return {
        organizations,
        namespaces: organizations.flatMap((organization, index) =>
          (namespacesByOrganization[index]?.items ?? []).map((namespace) => ({
            ...namespace,
            organization,
            organizationId: organization.id,
            organizationName: organization.name,
          }))
        ),
      };
    },
    ...cacheProfiles.reference,
  });
}

export function useAccessibleNamespaces() {
  const queryClient = useQueryClient();
  return useQuery(accessibleNamespacesOptions(queryClient));
}
