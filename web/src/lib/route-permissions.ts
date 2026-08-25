import type { QueryClient } from "@tanstack/react-query";

import { Action, ResourceType } from "@/lib/auth/permissions";
import { requirePermissionBeforeLoad } from "@/lib/auth/require-auth";
import {
  loadNamespaceHierarchy,
  loadOrganization,
  loadProjectHierarchy,
} from "@/lib/route-data";

export async function requireOrganizationAction(
  queryClient: QueryClient,
  organizationRef: string,
  action: string
) {
  const organization = await loadOrganization(queryClient, organizationRef);
  await requirePermissionBeforeLoad({
    queryClient,
    resourceType: ResourceType.Organization,
    action,
    resourceId: organization.id,
  });
  return organization;
}

export async function requireNamespaceAction(
  queryClient: QueryClient,
  organizationRef: string,
  namespaceRef: string,
  action: string
) {
  const hierarchy = await loadNamespaceHierarchy(
    queryClient,
    organizationRef,
    namespaceRef
  );
  await requirePermissionBeforeLoad({
    queryClient,
    resourceType: ResourceType.Namespace,
    action,
    resourceId: hierarchy.namespace.id,
  });
  return hierarchy;
}

export async function requireProjectAction(
  queryClient: QueryClient,
  organizationRef: string,
  namespaceRef: string,
  projectKey: string,
  action: string
) {
  const hierarchy = await loadProjectHierarchy(
    queryClient,
    organizationRef,
    namespaceRef,
    projectKey
  );
  await requirePermissionBeforeLoad({
    queryClient,
    resourceType: ResourceType.Project,
    action,
    resourceId: hierarchy.project.id,
  });
  return hierarchy;
}

export { Action, ResourceType };
