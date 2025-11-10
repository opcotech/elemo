import type { Page } from "@playwright/test";

/**
 * Wait for page to be ready by waiting for DOM content and network to be idle.
 * This is a more reliable alternative to waitForLoadState("networkidle").
 */
export async function waitForPageLoad(page: Page): Promise<void> {
  await page.waitForLoadState("domcontentloaded");
  try {
    await page.waitForLoadState("networkidle");
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
    timeout: options?.timeout ?? 5000,
  });
  await waitForPageLoad(page);
}
