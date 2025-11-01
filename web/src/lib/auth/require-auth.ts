import { redirect } from "@tanstack/react-router";

import { can } from "./permissions";
import { hasValidSession } from "./session";

import { withResourceType } from "@/hooks/use-permissions";
import type { ResourceType } from "@/hooks/use-permissions";
import { v1PermissionResourceGet } from "@/lib/api";
import type { PermissionKind } from "@/lib/api";

export function requireAuthBeforeLoad(ctx: { location: { href: string } }) {
  // Only run client-side to avoid SSR issues
  if (typeof window === "undefined") {
    return;
  }

  // Use a try-catch to handle any cookie/storage access issues
  try {
    if (!hasValidSession()) {
      throw redirect({
        to: "/login",
        search: {
          redirect: ctx.location.href,
        },
      });
    }
  } catch (error) {
    console.warn("Session check failed in route guard:", error);
  }
}

export function redirectIfAuthenticated() {
  if (hasValidSession()) {
    // If there's a redirect parameter, use it, otherwise go to dashboard
    const urlParams = new URLSearchParams(window.location.search);
    const redirectTo = urlParams.get("redirect");

    // Only redirect if the redirect target is not the login page itself
    if (redirectTo && !redirectTo.includes("/login")) {
      throw redirect({
        to: redirectTo,
      });
    } else {
      throw redirect({
        to: "/dashboard",
      });
    }
  }
}

/**
 * A permission requirement specifies a resource type, permission kind, and optionally a resource ID.
 */
export interface PermissionRequirement {
  resourceType: ResourceType;
  permissionKind: PermissionKind;
  resourceId?:
    | string
    | ((ctx: {
        location: { href: string };
        params?: Record<string, string>;
      }) => string | undefined);
}

/**
 * Creates a beforeLoad function that requires specific permissions.
 * First checks authentication, then checks for the required permissions.
 * When multiple requirements are provided, ALL must pass (AND condition).
 *
 * @param requirements - A single permission requirement or an array of requirements
 * @returns An async function that can be used as a beforeLoad handler
 */
export function requirePermissionBeforeLoad(
  requirements: PermissionRequirement | PermissionRequirement[]
): (ctx: {
  location: { href: string };
  params?: Record<string, string>;
}) => Promise<void> {
  return async (ctx) => {
    requireAuthBeforeLoad(ctx);

    // Normalize single requirement to array
    const requirementsArray = Array.isArray(requirements)
      ? requirements
      : [requirements];

    // Check all permissions (AND condition)
    for (const requirement of requirementsArray) {
      // Resolve resource ID if it's a function
      const resolvedResourceId =
        typeof requirement.resourceId === "function"
          ? requirement.resourceId(ctx)
          : requirement.resourceId;

      // Construct the resource ID string for permission checking
      const permissionResourceId = withResourceType(
        requirement.resourceType,
        resolvedResourceId
      );

      try {
        const response = await v1PermissionResourceGet({
          path: {
            resourceId: permissionResourceId,
          },
        });

        const permissions = response.data;
        if (!can(permissions, requirement.permissionKind)) {
          throw redirect({
            to: "/permission-denied",
          });
        }
      } catch (error) {
        // If it's already a redirect, rethrow it
        if (error && typeof error === "object" && "to" in error) {
          throw error;
        }
        // On error fetching permissions, deny access
        throw redirect({
          to: "/permission-denied",
        });
      }
    }
  };
}
