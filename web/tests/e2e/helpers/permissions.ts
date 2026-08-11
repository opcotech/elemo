import type { Page } from "@playwright/test";

import { waitForAPIResponse } from "./api";

/**
 * Wait for a permission API response.
 *
 * Prefer UI readiness markers after navigation. Use this only when the waiter
 * is registered before the request can complete (for example with Promise.all
 * around a click/navigation that triggers the permission fetch).
 */
export async function waitForPermissionsLoad(
  page: Page,
  resourceId?: string,
  options?: { timeout?: number; requireOk?: boolean }
): Promise<void> {
  const pattern = resourceId
    ? new RegExp(
        `/v1/permissions/resources/${resourceId.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}`
      )
    : /\/v1\/permissions\/resources\//;

  await waitForAPIResponse(page, pattern, options);
}
