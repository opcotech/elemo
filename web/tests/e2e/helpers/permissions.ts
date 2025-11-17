import type { Page } from "@playwright/test";

import { waitForAPIResponse } from "./api";

/**
 * Wait for permission API calls to complete.
 * This waits for the permissions endpoint that's commonly used across the app.
 */
export async function waitForPermissionsLoad(
  page: Page,
  resourceId?: string
): Promise<void> {
  const pattern = resourceId
    ? new RegExp(
        `/v1/permissions/resources/${resourceId.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}`
      )
    : /\/v1\/permissions\/resources\//;

  try {
    await waitForAPIResponse(page, pattern);
  } catch {}
}
