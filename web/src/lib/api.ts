import type { z } from "zod";

import { isEmpty } from "./utils";

import { config } from "@/config";
import { getAccessToken } from "@/lib/auth/session";
import { client } from "@/lib/client/client.gen";

let isClientConfigured = false;

/**
 * Configure the shared API client. Safe to call multiple times.
 * Must run before API requests in the app; deferred from module load so
 * Playwright type-imports of this module do not assert Vite env vars.
 */
export function ensureApiClientConfigured(): void {
  if (isClientConfigured) {
    return;
  }

  client.setConfig({
    baseUrl: config.auth().apiBaseUrl,
    cache: "no-store",
    auth: async () => {
      // Do not attempt to read localStorage or run refresh logic on the server.
      if (typeof window === "undefined") {
        return undefined;
      }

      try {
        let token = await getAccessToken();

        // If no token available, try to refresh
        if (!token) {
          const { tokenRefreshService } =
            await import("@/lib/auth/token-refresh-service");
          if (!tokenRefreshService.isRefreshInProgress()) {
            await tokenRefreshService.forceRefresh();
            token = await getAccessToken();
          }
        }

        return token || undefined;
      } catch (error) {
        console.error("Failed to get access token:", error);
        return undefined;
      }
    },
  });

  isClientConfigured = true;
}

function hasApiBaseUrl(): boolean {
  try {
    if (
      typeof import.meta !== "undefined" &&
      import.meta.env?.VITE_API_BASE_URL
    ) {
      return true;
    }
  } catch {
    // import.meta unavailable during some Node/CJS analysis paths
  }
  return Boolean(
    typeof process !== "undefined" && process.env?.VITE_API_BASE_URL
  );
}

// Configure when Vite/env is present (app + SSR). Skip during Playwright analysis.
if (hasApiBaseUrl()) {
  ensureApiClientConfigured();
}

export * from "@/lib/client/@tanstack/react-query.gen";
export * from "@/lib/client/client.gen";
export * from "@/lib/client/sdk.gen";
export * from "@/lib/client/types.gen";

/**
 * Normalizes form data by converting empty strings to undefined for optional fields.
 * This ensures that empty optional fields are omitted from the request body rather than sent as null.
 *
 * @param schema - The Zod schema to check which fields are optional
 * @param data - The data to normalize
 * @returns The normalized data with empty strings converted to undefined for optional fields
 */
export function normalizeData<T extends Record<string, any>>(
  schema: z.ZodObject<any>,
  data: T
): Partial<T> {
  const normalizedData: Partial<T> = { ...data };

  for (const [key, value] of Object.entries(data)) {
    if (schema.shape[key as keyof T]?.isOptional() && isEmpty(value)) {
      delete normalizedData[key as keyof T];
    }
  }

  return normalizedData;
}

/**
 * Checks if an error represents a permission denied (403) response.
 *
 * @param error - The error object to check
 * @returns true if the error is a 403 permission denied error, false otherwise
 */
export function isPermissionDenied(error: unknown): boolean {
  if (!error) return false;

  if (
    typeof error === "object" &&
    error !== null &&
    "status" in error &&
    error.status === 403
  ) {
    return true;
  }

  if (
    typeof error === "object" &&
    error !== null &&
    "response" in error &&
    typeof error.response === "object" &&
    error.response !== null &&
    "status" in error.response &&
    error.response.status === 403
  ) {
    return true;
  }

  if (
    typeof error === "object" &&
    error !== null &&
    "message" in error &&
    typeof error.message === "string"
  ) {
    const message = error.message.toLowerCase();
    return message.includes("403") || message.includes("forbidden");
  }

  return false;
}

/**
 * Checks if an error represents a not found (404) response.
 *
 * @param error - The error object to check
 * @returns true if the error is a 404 not found error, false otherwise
 */
export function isNotFound(error: unknown): boolean {
  if (!error) return false;

  if (
    typeof error === "object" &&
    error !== null &&
    "status" in error &&
    error.status === 404
  ) {
    return true;
  }

  if (
    typeof error === "object" &&
    error !== null &&
    "response" in error &&
    typeof error.response === "object" &&
    error.response !== null &&
    "status" in error.response &&
    error.response.status === 404
  ) {
    return true;
  }

  if (
    typeof error === "object" &&
    error !== null &&
    "message" in error &&
    typeof error.message === "string"
  ) {
    const message = error.message.toLowerCase();
    return message.includes("404") || message.includes("not found");
  }

  return false;
}
