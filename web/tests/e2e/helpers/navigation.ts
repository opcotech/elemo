import type { Page } from "@playwright/test";

/**
 * Wait for page to be ready by waiting for DOM content and network to be idle.
 * This is a more reliable alternative to waitForLoadState("networkidle").
 */
export async function waitForPageLoad(page: Page): Promise<void> {
  await page.waitForLoadState("domcontentloaded");
  try {
    await page.waitForLoadState("networkidle", { timeout: 5000 });
  } catch {}
}

/**
 * Wait for a specific API response to complete.
 * @param page - Playwright page object
 * @param urlPattern - URL pattern to match (string or RegExp)
 * @param options - Optional timeout and other options
 */
export async function waitForAPIResponse(
  page: Page,
  urlPattern: string | RegExp,
  options?: { timeout?: number }
): Promise<void> {
  const timeout = options?.timeout ?? 10000;
  await page.waitForResponse(
    (response) => {
      const url = response.url();
      if (typeof urlPattern === "string") {
        return url.includes(urlPattern);
      }
      return urlPattern.test(url);
    },
    { timeout }
  );
}

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
    await waitForAPIResponse(page, pattern, { timeout: 3000 });
  } catch {}
}

/**
 * Navigate to a URL and wait for the page to be ready.
 * @param page - Playwright page object
 * @param url - URL to navigate to
 * @param options - Optional options for navigation and waiting
 */
export async function navigateAndWait(
  page: Page,
  url: string,
  options?: {
    waitUntil?: "load" | "domcontentloaded" | "networkidle" | "commit";
    timeout?: number;
  }
): Promise<void> {
  await page.goto(url, {
    waitUntil: options?.waitUntil ?? "domcontentloaded",
    timeout: options?.timeout ?? 30000,
  });
  await waitForPageLoad(page);
}
