import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

/**
 * Wait for the current document and an optional route-specific UI marker.
 * Persistent connections and background queries make network-idle unsuitable
 * as an application-readiness signal.
 */
export async function waitForPageLoad(
  page: Page,
  ready?: Locator
): Promise<void> {
  await page.waitForLoadState("domcontentloaded");
  await expect(ready ?? page.locator("body")).toBeVisible();
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
    waitUntil?: "load" | "domcontentloaded" | "commit";
    timeout?: number;
    ready?: Locator;
  }
): Promise<void> {
  const waitUntil = options?.waitUntil ?? "domcontentloaded";
  const timeout = options?.timeout ?? 15_000;

  // WebKit (and occasionally Firefox) can abort an in-flight document load when
  // cookies are cleared or another navigation starts; retry once.
  for (let attempt = 0; attempt < 2; attempt++) {
    try {
      await page.goto(url, { waitUntil, timeout });
      break;
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      const interrupted =
        message.includes("Frame load interrupted") ||
        message.includes("interrupted by another navigation") ||
        message.includes("NS_BINDING_ABORTED") ||
        message.includes("NS_ERROR_FAILURE");
      if (!interrupted || attempt === 1) {
        throw error;
      }
    }
  }
  await waitForPageLoad(page, options?.ready);
}
