import { useParams, useRouterState } from "@tanstack/react-router";
import { useMemo } from "react";

export type NavigationContextType = "namespace" | "project" | "global";

export interface NavigationContext {
  type: NavigationContextType;
  namespaceId?: string;
  projectId?: string;
  organizationId?: string;
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

  return useMemo(() => {
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
        type: "global",
        organizationId: params.organizationId,
      };
    }

    return {
      type: "global",
    };
  }, [pathname, params.organizationId, params.namespaceId, params.projectId]);
}
