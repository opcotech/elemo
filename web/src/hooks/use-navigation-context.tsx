import { useParams, useRouterState } from "@tanstack/react-router";
import { useMemo } from "react";

import { identityFromMatches } from "@/lib/route-identity";

export type NavigationContextType =
  "organization" | "namespace" | "project" | "global";

export interface NavigationContext {
  type: NavigationContextType;
  organizationSlug?: string;
  namespaceSlug?: string;
  projectKey?: string;
  organizationId?: string;
  namespaceId?: string;
  projectId?: string;
}

export interface NavigationContextParams {
  organizationSlug?: string;
  namespaceSlug?: string;
  projectKey?: string;
  organizationId?: string;
  namespaceId?: string;
  projectId?: string;
}

/**
 * Derives operational sidebar context from a pathname, URL identity, and
 * resolved xids from loader data. Settings admin routes share param names
 * but must not replace operational context.
 */
export function resolveNavigationContext(
  pathname: string,
  params: NavigationContextParams
): NavigationContext {
  if (pathname.startsWith("/settings")) {
    return { type: "global" };
  }

  if (params.projectKey) {
    return {
      type: "project",
      organizationSlug: params.organizationSlug,
      namespaceSlug: params.namespaceSlug,
      projectKey: params.projectKey,
      organizationId: params.organizationId,
      namespaceId: params.namespaceId,
      projectId: params.projectId,
    };
  }

  if (params.namespaceSlug) {
    return {
      type: "namespace",
      organizationSlug: params.organizationSlug,
      namespaceSlug: params.namespaceSlug,
      organizationId: params.organizationId,
      namespaceId: params.namespaceId,
    };
  }

  if (params.organizationSlug) {
    return {
      type: "organization",
      organizationSlug: params.organizationSlug,
      organizationId: params.organizationId,
    };
  }

  return {
    type: "global",
  };
}

/**
 * Derives the sidebar context from the router's typed parameters and loader
 * data. URL identity comes from params; xids come from loaders only.
 */
export function useNavigationContext(): NavigationContext {
  const params = useParams({ strict: false });
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });
  const matches = useRouterState({
    select: (state) => state.matches,
  });

  const resolved = useMemo(() => identityFromMatches(matches), [matches]);

  return useMemo(
    () =>
      resolveNavigationContext(pathname, {
        organizationSlug: params.organizationSlug,
        namespaceSlug: params.namespaceSlug,
        projectKey: params.projectKey,
        organizationId: resolved.organizationId,
        namespaceId: resolved.namespaceId,
        projectId: resolved.projectId,
      }),
    [
      pathname,
      params.organizationSlug,
      params.namespaceSlug,
      params.projectKey,
      resolved.organizationId,
      resolved.namespaceId,
      resolved.projectId,
    ]
  );
}
