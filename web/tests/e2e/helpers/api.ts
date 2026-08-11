import type { Page } from "@playwright/test";

/**
 * Wait for a specific API response to complete.
 * @param page - Playwright page object
 * @param urlPattern - URL pattern to match (string or RegExp)
 * @param options - Optional timeout and other options
 */
export async function waitForAPIResponse(
  page: Page,
  urlPattern: string | RegExp,
  options?: { timeout?: number; requireOk?: boolean }
): Promise<void> {
  const timeout = options?.timeout ?? 5000;
  const response = await page.waitForResponse(
    (candidate) => {
      const url = candidate.url();
      if (typeof urlPattern === "string") {
        return url.includes(urlPattern);
      }
      urlPattern.lastIndex = 0;
      return urlPattern.test(url);
    },
    { timeout }
  );

  if ((options?.requireOk ?? true) && !response.ok()) {
    throw new Error(
      `API readiness request failed: ${response.status()} ${response.url()}`
    );
  }
}
