import type { z } from "zod";

import { config } from "@/config";
import { getAccessToken } from "@/lib/auth/session";
import { tokenRefreshService } from "@/lib/auth/token-refresh-service";
import { client } from "@/lib/client/client.gen";

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
      if (!token && !tokenRefreshService.isRefreshInProgress()) {
        await tokenRefreshService.forceRefresh();
        token = await getAccessToken();
      }

      return token || undefined;
    } catch (error) {
      console.error("Failed to get access token:", error);
      return undefined;
    }
  },
});

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

  function isEmpty(value: any) {
    return (
      value === null ||
      value === undefined ||
      (typeof value === "string" && value.trim() === "")
    );
  }

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
