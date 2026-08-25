import type { QueryClient } from "@tanstack/react-query";

import {
  v1NamespaceGetOptions,
  v1NamespacesProjectsKeyGetOptions,
  v1ProjectGetOptions,
} from "@/lib/api/query-options";
import { namespaceRefPath, projectIdPath } from "@/lib/api/refs";
import type { AccessibleNamespace, Project } from "@/lib/api/types";
import { Action, ResourceType } from "@/lib/auth/permissions";
import { requireResourcePermission } from "@/lib/entity-context";
import { withRouteErrors } from "@/lib/route-errors";
import {
  requireNamespaceSlug,
  requireOrganizationSlug,
  requireProjectKey,
} from "@/lib/route-identity";

export async function loadAccessibleNamespace(
  queryClient: QueryClient,
  organizationSlug: string,
  namespaceSlug: string
): Promise<AccessibleNamespace> {
  const namespace = await queryClient.fetchQuery(
    v1NamespaceGetOptions({
      path: namespaceRefPath(organizationSlug, namespaceSlug),
    })
  );
  queryClient.setQueryData(
    v1NamespaceGetOptions({
      path: namespaceRefPath(namespace.organization.id, namespace.id),
    }).queryKey,
    namespace
  );
  return namespace;
}

export async function loadNamespaceOperationalContext(
  queryClient: QueryClient,
  organizationSlug: string,
  namespaceSlug: string
) {
  requireOrganizationSlug(organizationSlug);
  requireNamespaceSlug(namespaceSlug);

  return withRouteErrors(async () => {
    const namespace = await loadAccessibleNamespace(
      queryClient,
      organizationSlug,
      namespaceSlug
    );

    return {
      namespace,
      organization: namespace.organization,
    };
  }, "redirect");
}

export async function loadProjectByKey(
  queryClient: QueryClient,
  organizationRef: string,
  namespaceRef: string,
  projectKey: string
): Promise<Project> {
  const project = await queryClient.fetchQuery(
    v1NamespacesProjectsKeyGetOptions({
      path: { ...namespaceRefPath(organizationRef, namespaceRef), projectKey },
    })
  );
  queryClient.setQueryData(
    v1ProjectGetOptions({ path: projectIdPath(project.id) }).queryKey,
    project
  );
  return project;
}

export async function loadProjectOperationalContext(
  queryClient: QueryClient,
  organizationSlug: string,
  namespaceSlug: string,
  projectKey: string
) {
  requireOrganizationSlug(organizationSlug);
  requireNamespaceSlug(namespaceSlug);
  requireProjectKey(projectKey);

  return withRouteErrors(async () => {
    const namespacePromise = loadAccessibleNamespace(
      queryClient,
      organizationSlug,
      namespaceSlug
    );
    const projectPromise = loadProjectByKey(
      queryClient,
      organizationSlug,
      namespaceSlug,
      projectKey
    );

    const [namespace, project] = await Promise.all([
      namespacePromise,
      projectPromise,
    ]);

    await requireResourcePermission(
      queryClient,
      ResourceType.Project,
      project.id,
      Action.ProjectRead
    );

    return {
      namespace,
      organization: namespace.organization,
      project,
    };
  }, "redirect");
}
