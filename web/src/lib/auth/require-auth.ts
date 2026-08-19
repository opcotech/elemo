import type { QueryClient } from "@tanstack/react-query";
import { redirect } from "@tanstack/react-router";

import { can } from "./permissions";
import type { ResourceType } from "./permissions";

import { loadResourcePermissions } from "@/lib/entity-context";

export async function redirectIfAuthenticated() {
  const { currentSessionFn } = await import("./functions");
  if (await currentSessionFn()) {
    throw redirect({
      to: "/",
    });
  }
}

export async function requirePermissionBeforeLoad({
  queryClient,
  resourceType,
  action,
  resourceId,
}: {
  queryClient: QueryClient;
  resourceType: ResourceType;
  action: string;
  resourceId?: string;
}) {
  const permissions = await loadResourcePermissions(
    queryClient,
    resourceType,
    resourceId
  );

  if (!can(permissions, action)) {
    throw redirect({
      to: "/permission-denied",
    });
  }

  return permissions;
}
