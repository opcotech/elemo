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

export function normalizeData<T extends Record<string, any>>(
  schema: z.ZodObject<any>,
  data: T
): T {
  const normalizedData: Partial<T> = { ...data };

  function isEmpty(value: any) {
    return value === null || (value !== undefined && value.trim?.() === "");
  }

  for (const [key, value] of Object.entries(data)) {
    if (schema.shape[key as keyof T]?.isOptional() && isEmpty(value)) {
      normalizedData[key as keyof T] = null as T[keyof T];
    }
  }

  return normalizedData as T;
}
