import { useParams, useRouterState } from "@tanstack/react-router";
import { useMemo } from "react";

export type NavigationContextType =
  "organization" | "namespace" | "project" | "global";

export interface NavigationContext {
  type: NavigationContextType;
  namespaceId?: string;
  projectId?: string;
  organizationId?: string;
}

export interface NavigationContextParams {
  organizationId?: string;
  namespaceId?: string;
  projectId?: string;
}

/**
 * Derives operational sidebar context from a pathname and merged route params.
 * Settings admin routes share param names but must not replace operational context.
 */
export function resolveNavigationContext(
  pathname: string,
  params: NavigationContextParams
): NavigationContext {
  if (pathname.startsWith("/settings")) {
    return { type: "global" };
  }

  if (params.projectId) {
    return {
      type: "project",
      organizationId: params.organizationId,
      namespaceId: params.namespaceId,
      projectId: params.projectId,
    };
  }

  if (params.namespaceId) {
    return {
      type: "namespace",
      organizationId: params.organizationId,
      namespaceId: params.namespaceId,
    };
  }

  if (params.organizationId) {
    return {
      type: "organization",
      organizationId: params.organizationId,
    };
  }

  return {
    type: "global",
  };
}

/**
 * Derives the sidebar context from the router's typed, merged parameters.
 * Settings admin routes share param names but must not replace operational context.
 */
export function useNavigationContext(): NavigationContext {
  const params = useParams({ strict: false });
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  return useMemo(
    () =>
      resolveNavigationContext(pathname, {
        organizationId: params.organizationId,
        namespaceId: params.namespaceId,
        projectId: params.projectId,
      }),
    [pathname, params.organizationId, params.namespaceId, params.projectId]
  );
}
