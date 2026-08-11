import { notFound, redirect } from "@tanstack/react-router";

import { isNotFound, isPermissionDenied } from "@/lib/api/errors";

export const accessDeniedRouteData = { accessDenied: true } as const;

export type AccessDeniedRouteData = typeof accessDeniedRouteData;

export function isAccessDeniedRouteData(
  data: unknown
): data is AccessDeniedRouteData {
  return (
    !!data &&
    typeof data === "object" &&
    "accessDenied" in data &&
    data.accessDenied === true
  );
}

export function withRouteErrors<T>(load: () => Promise<T>): Promise<T>;
export function withRouteErrors<T>(
  load: () => Promise<T>,
  permissionDenied: "redirect"
): Promise<T>;
export function withRouteErrors<T>(
  load: () => Promise<T>,
  permissionDenied: "data"
): Promise<T | AccessDeniedRouteData>;
export async function withRouteErrors<T>(
  load: () => Promise<T>,
  permissionDenied: "throw" | "redirect" | "data" = "throw"
): Promise<T | AccessDeniedRouteData> {
  try {
    return await load();
  } catch (error) {
    if (isNotFound(error)) {
      throw notFound();
    }
    if (isPermissionDenied(error)) {
      if (permissionDenied === "redirect") {
        throw redirect({ to: "/permission-denied" });
      }
      if (permissionDenied === "data") {
        return accessDeniedRouteData;
      }
    }
    throw error;
  }
}
