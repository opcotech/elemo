import type { QueryClient } from "@tanstack/react-query";
import { notFound } from "@tanstack/react-router";

import { accessibleNamespacesOptions } from "@/lib/api/accessible-namespaces";
import { v1ProjectGetOptions } from "@/lib/api/query-options";
import { ResourceType } from "@/lib/auth/permissions";
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

    await requireResourcePermission(
      queryClient,
      ResourceType.Namespace,
      namespaceId
    );

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
    const { namespace, organization } = await loadNamespaceOperationalContext(
      queryClient,
      namespaceId
    );

    if (!namespace.projects.some((project) => project.id === projectId)) {
      throw notFound();
    }

    const project = await queryClient.fetchQuery(
      v1ProjectGetOptions({ path: { id: projectId } })
    );
    await requireResourcePermission(
      queryClient,
      ResourceType.Project,
      projectId
    );

    return { namespace, organization, project };
  }, "redirect");
}
