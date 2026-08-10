import { useRouter } from "@tanstack/react-router";
import { useMemo } from "react";

export type NavigationContextType = "namespace" | "project" | "global";

export interface NavigationContext {
  type: NavigationContextType;
  namespaceId?: string;
  projectId?: string;
  organizationId?: string;
}

const RESERVED_SEGMENTS = new Set(["new", "edit"]);

/**
 * Hook to detect the current navigation context based on the route.
 * Returns the context type and relevant IDs extracted from route params.
 */
export function useNavigationContext(): NavigationContext {
  const router = useRouter();
  const location = router.state.location;

  return useMemo(() => {
    const pathname = location.pathname;

    // Pattern: /settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId
    const namespaceProjectMatch = pathname.match(
      /\/settings\/organizations\/([^/]+)\/namespaces\/([^/]+)\/projects\/([^/]+)/
    );
    if (
      namespaceProjectMatch &&
      !RESERVED_SEGMENTS.has(namespaceProjectMatch[2])
    ) {
      const projectId = RESERVED_SEGMENTS.has(namespaceProjectMatch[3])
        ? undefined
        : namespaceProjectMatch[3];

      return {
        // Keep namespace type so existing sidebar navigation stays correct for
        // nested settings project routes (project sidebar paths are unfinished).
        type: "namespace",
        organizationId: namespaceProjectMatch[1],
        namespaceId: namespaceProjectMatch[2],
        projectId,
      };
    }

    // Pattern: /settings/organizations/$organizationId/namespaces/$namespaceId
    const namespaceMatch = pathname.match(
      /\/settings\/organizations\/([^/]+)\/namespaces\/([^/]+)/
    );
    if (namespaceMatch && !RESERVED_SEGMENTS.has(namespaceMatch[2])) {
      return {
        type: "namespace",
        organizationId: namespaceMatch[1],
        namespaceId: namespaceMatch[2],
      };
    }

    // Pattern: /projects/$projectId (legacy / non-settings project routes)
    const projectMatch = pathname.match(/\/projects\/([^/]+)/);
    if (projectMatch && !RESERVED_SEGMENTS.has(projectMatch[1])) {
      return {
        type: "project",
        projectId: projectMatch[1],
      };
    }

    // Pattern: /settings/organizations/$organizationId (detail/edit/roles/etc.)
    const organizationMatch = pathname.match(
      /\/settings\/organizations\/([^/]+)/
    );
    if (organizationMatch && !RESERVED_SEGMENTS.has(organizationMatch[1])) {
      return {
        type: "global",
        organizationId: organizationMatch[1],
      };
    }

    // Default to global context
    return {
      type: "global",
    };
  }, [location.pathname]);
}
