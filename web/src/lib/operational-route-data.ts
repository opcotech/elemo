import type { QueryClient } from "@tanstack/react-query";
import { notFound } from "@tanstack/react-router";

import { accessibleNamespacesOptions } from "@/lib/api/accessible-namespaces";
import { collectListedPage, cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1NamespacesProjectsGetOptions,
  v1ProjectGetOptions,
} from "@/lib/api/query-options";
import { Action, ResourceType } from "@/lib/auth/permissions";
import { requireResourcePermission } from "@/lib/entity-context";
import { withRouteErrors } from "@/lib/route-errors";

export async function loadNamespaceOperationalContext(
  queryClient: QueryClient,
  namespaceId: string
) {
  return withRouteErrors(async () => {
    const accessibleWorkspace = await queryClient.fetchQuery(
      accessibleNamespacesOptions(queryClient)
    );
    const accessibleNamespace = accessibleWorkspace.namespaces.find(
      (item) => item.id === namespaceId
    );

    if (!accessibleNamespace) {
      throw notFound();
    }

    return {
      namespace: accessibleNamespace,
      organization: accessibleNamespace.organization,
    };
  }, "redirect");
}

export async function loadProjectOperationalContext(
  queryClient: QueryClient,
  namespaceId: string,
  projectId: string
) {
  return withRouteErrors(async () => {
    const accessibleWorkspace = await queryClient.fetchQuery(
      accessibleNamespacesOptions(queryClient)
    );
    const accessibleNamespace = accessibleWorkspace.namespaces.find(
      (item) => item.id === namespaceId
    );

    if (!accessibleNamespace) {
      throw notFound();
    }

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

    const project = await queryClient.fetchQuery(
      v1ProjectGetOptions({ path: { id: projectId } })
    );
    await requireResourcePermission(
      queryClient,
      ResourceType.Project,
      projectId,
      Action.ProjectRead
    );

    return {
      namespace: accessibleNamespace,
      organization: accessibleNamespace.organization,
      project,
    };
  }, "redirect");
}
