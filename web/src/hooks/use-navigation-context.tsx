import { useRouter } from "@tanstack/react-router";
import { useMemo } from "react";

export type NavigationContextType = "namespace" | "project" | "global";

export interface NavigationContext {
  type: NavigationContextType;
  namespaceId?: string;
  projectId?: string;
  organizationId?: string;
}

/**
 * Hook to detect the current navigation context based on the route.
 * Returns the context type and relevant IDs extracted from route params.
 */
export function useNavigationContext(): NavigationContext {
  const router = useRouter();
  const location = router.state.location;

  return useMemo(() => {
    const pathname = location.pathname;

    // Check for namespace context
    // Pattern: /settings/organizations/$organizationId/namespaces/$namespaceId
    const namespaceMatch = pathname.match(
      /\/settings\/organizations\/([^/]+)\/namespaces\/([^/]+)/
    );
    if (namespaceMatch) {
      return {
        type: "namespace",
        organizationId: namespaceMatch[1],
        namespaceId: namespaceMatch[2],
      };
    }

    // Check for project context
    // Pattern: /projects/$projectId or /namespaces/$namespaceId/projects/$projectId
    const projectMatch = pathname.match(/\/projects\/([^/]+)/);
    if (projectMatch) {
      return {
        type: "project",
        projectId: projectMatch[1],
      };
    }

    // Check for project within namespace context
    // Pattern: /settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId
    const namespaceProjectMatch = pathname.match(
      /\/settings\/organizations\/([^/]+)\/namespaces\/([^/]+)\/projects\/([^/]+)/
    );
    if (namespaceProjectMatch) {
      return {
        type: "project",
        organizationId: namespaceProjectMatch[1],
        namespaceId: namespaceProjectMatch[2],
        projectId: namespaceProjectMatch[3],
      };
    }

    // Default to global context
    return {
      type: "global",
    };
  }, [location.pathname]);
}
