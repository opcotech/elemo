import { redirect } from "@tanstack/react-router";

import { can } from "./permissions";
import { getAccessToken, hasValidSession } from "./session";
import { tokenRefreshService } from "./token-refresh-service";

import { withResourceType } from "@/hooks/use-permissions";
import type { ResourceType } from "@/hooks/use-permissions";
import { isPermissionDenied, v1PermissionResourceGet } from "@/lib/api";
import type { PermissionKind } from "@/lib/api";

/**
 * Context passed to route beforeLoad handlers
 */
interface RouteContext {
  location: { href: string };
  params?: Record<string, string>;
}

/**
 * Maximum time to wait for token refresh (in milliseconds)
 */
const TOKEN_REFRESH_MAX_WAIT = 5000;
const TOKEN_REFRESH_POLL_INTERVAL = 100;

/**
 * Checks if an error represents an unauthorized (401) response.
 */
function isUnauthorized(error: unknown): boolean {
  if (!error || typeof error !== "object") {
    return false;
  }

  // Check direct status property
  if ("status" in error && error.status === 401) {
    return true;
  }

  // Check nested response status
  if (
    "response" in error &&
    typeof error.response === "object" &&
    error.response !== null &&
    "status" in error.response &&
    error.response.status === 401
  ) {
    return true;
  }

  // Check error message
  if ("message" in error && typeof error.message === "string") {
    const message = error.message.toLowerCase();
    return message.includes("401") || message.includes("unauthorized");
  }

  return false;
}

/**
 * Checks if an error is already a redirect object.
 */
function isRedirect(error: unknown): error is { to: string } {
  return (
    typeof error === "object" &&
    error !== null &&
    "to" in error &&
    typeof error.to === "string"
  );
}

/**
 * Creates a redirect to the login page with a return URL.
 */
function redirectToLogin(currentUrl: string): never {
  throw redirect({
    to: "/login",
    search: {
      redirect: currentUrl,
    },
  });
}

/**
 * Creates a redirect to the permission denied page.
 */
function redirectToPermissionDenied(): never {
  throw redirect({
    to: "/permission-denied",
  });
}

/**
 * Ensures an access token is available, refreshing if necessary.
 * Waits for in-progress refreshes and attempts a new refresh if needed.
 *
 * @param currentUrl - The current URL for redirect purposes
 * @throws Redirects to login if token cannot be obtained
 */
async function ensureAccessToken(currentUrl: string): Promise<string> {
  let accessToken = await getAccessToken();

  // Wait for in-progress refresh if needed
  if (!accessToken && tokenRefreshService.isRefreshInProgress()) {
    const startTime = Date.now();
    while (!accessToken && Date.now() - startTime < TOKEN_REFRESH_MAX_WAIT) {
      await new Promise((resolve) =>
        setTimeout(resolve, TOKEN_REFRESH_POLL_INTERVAL)
      );
      accessToken = await getAccessToken();
    }
  }

  // Attempt to refresh if no token available
  if (!accessToken && !tokenRefreshService.isRefreshInProgress()) {
    try {
      await tokenRefreshService.forceRefresh();
      accessToken = await getAccessToken();
    } catch {
      redirectToLogin(currentUrl);
    }
  }

  // If still no token, redirect to login
  if (!accessToken) {
    redirectToLogin(currentUrl);
  }

  return accessToken;
}

/**
 * Resolves the resource ID from a requirement, handling function-based IDs.
 */
function resolveResourceId(
  requirement: PermissionRequirement,
  ctx: RouteContext
): string | undefined {
  return typeof requirement.resourceId === "function"
    ? requirement.resourceId(ctx)
    : requirement.resourceId;
}

/**
 * Checks if a user has the required permission for a resource.
 *
 * @param requirement - The permission requirement to check
 * @param ctx - The route context
 * @param currentUrl - The current URL for redirect purposes
 * @throws Redirects to login or permission denied on failure
 */
async function checkPermission(
  requirement: PermissionRequirement,
  ctx: RouteContext,
  currentUrl: string
): Promise<void> {
  const resolvedResourceId = resolveResourceId(requirement, ctx);
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
      redirectToPermissionDenied();
    }
  } catch (error) {
    handlePermissionCheckError(error, currentUrl);
  }
}

/**
 * Handles errors from permission checks, redirecting appropriately.
 *
 * @param error - The error from the permission check
 * @param currentUrl - The current URL for redirect purposes
 * @throws Redirects to login or permission denied, or allows page load for other errors
 */
function handlePermissionCheckError(error: unknown, currentUrl: string): void {
  // If it's already a redirect, rethrow it
  if (isRedirect(error)) {
    throw error;
  }

  // Handle permission denied (403)
  if (isPermissionDenied(error)) {
    redirectToPermissionDenied();
  }

  // Handle unauthorized (401)
  if (isUnauthorized(error)) {
    redirectToLogin(currentUrl);
  }

  // For other errors (network, etc.), allow the page to load
  // The component can handle permission checks client-side
  // This prevents false positives on page refresh when network/auth isn't ready
  console.warn(
    "Permission check failed in beforeLoad, allowing page to load:",
    error
  );
  // Don't throw redirect - let the page load and handle permissions in component
}

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
  resourceId?: string | ((ctx: RouteContext) => string | undefined);
}

/**
 * Creates a beforeLoad function that requires specific permissions.
 * First checks authentication, then checks for the required permissions.
 * When multiple requirements are provided, ALL must pass (AND condition).
 *
 * @param requirements - A single permission requirement or an array of requirements
 * @returns An async function that can be used as a beforeLoad handler
 *
 * @example Single permission:
 * ```ts
 * beforeLoad: requirePermissionBeforeLoad({
 *   resourceType: ResourceType.Organization,
 *   permissionKind: "write",
 * })
 * ```
 *
 * @example Multiple permissions (AND condition):
 * ```ts
 * beforeLoad: requirePermissionBeforeLoad([
 *   {
 *     resourceType: ResourceType.Organization,
 *     permissionKind: "read",
 *     resourceId: (ctx) => ctx.params?.organizationId,
 *   },
 *   {
 *     resourceType: ResourceType.Project,
 *     permissionKind: "write",
 *   },
 * ])
 * ```
 */
export function requirePermissionBeforeLoad(
  requirements: PermissionRequirement | PermissionRequirement[]
): (ctx: RouteContext) => Promise<void> {
  return async (ctx) => {
    // Only run client-side to avoid SSR issues
    if (typeof window === "undefined") {
      return;
    }

    requireAuthBeforeLoad(ctx);

    // Ensure we have an access token before making permission checks
    await ensureAccessToken(ctx.location.href);

    // Normalize single requirement to array
    const requirementsArray = Array.isArray(requirements)
      ? requirements
      : [requirements];

    // Check all permissions (AND condition)
    for (const requirement of requirementsArray) {
      await checkPermission(requirement, ctx, ctx.location.href);
    }
  };
}
