import { expect } from "@playwright/test";
import type { Page } from "@playwright/test";

/**
 * Helper functions for working with toast notifications.
 * Provides consistent ways to wait for and verify toast messages.
 */

/**
 * Wait for a toast notification to appear with specific text.
 * @param page - Playwright page object
 * @param text - Text to search for in the toast (can be partial)
 * @param options - Optional timeout
 */
export async function waitForToast(
  page: Page,
  text: string,
  options?: { timeout?: number; exact?: boolean }
): Promise<void> {
  const timeout = options?.timeout ?? 10000;
  const exact = options?.exact ?? false;
  const titleLocator = page.locator(`[data-title="${text}"]`);
  const textLocator = page.getByText(text, { exact });
  try {
    await expect(titleLocator).toBeVisible({
      timeout: Math.min(timeout, 2000),
    });
    return;
  } catch {
    const count = await textLocator.count();
    if (count > 0) {
      await expect(textLocator.first()).toBeVisible({ timeout });
    } else {
      await page.waitForTimeout(500);
      await expect(textLocator.first()).toBeVisible({ timeout });
    }
  }
}

/**
 * Wait for a success toast to appear.
 * @param page - Playwright page object
 * @param text - Optional specific text to look for
 * @param options - Optional timeout
 */
export async function waitForSuccessToast(
  page: Page,
  text?: string,
  options?: { timeout?: number }
): Promise<void> {
  if (text) {
    await waitForToast(page, text, options);
  } else {
    await Promise.race([
      waitForToast(page, "successfully", options).catch(() => null),
      waitForToast(page, "created", options).catch(() => null),
      waitForToast(page, "updated", options).catch(() => null),
      waitForToast(page, "deleted", options).catch(() => null),
    ]);
  }
}

/**
 * Wait for an error toast to appear.
 * @param page - Playwright page object
 * @param text - Optional specific text to look for
 * @param options - Optional timeout
 */
export async function waitForErrorToast(
  page: Page,
  text?: string,
  options?: { timeout?: number }
): Promise<void> {
  if (text) {
    await waitForToast(page, text, options);
  } else {
    await expect(page.locator('[role="alert"]')).toBeVisible({
      timeout: options?.timeout ?? 10000,
    });
  }
}

/**
 * Check if a toast is visible.
 * @param page - Playwright page object
 * @param text - Text to search for
 * @returns True if toast is visible, false otherwise
 */
export async function isToastVisible(
  page: Page,
  text: string
): Promise<boolean> {
  return await page
    .getByText(text)
    .isVisible()
    .catch(() => false);
}
