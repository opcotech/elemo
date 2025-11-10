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
  const timeout = options?.timeout ?? 5000;
  const exact = options?.exact ?? false;
  const searchValue = buildTextMatcher(text, exact);
  const toastLocator = page.locator("[data-sonner-toast]").filter({
    hasText: searchValue,
  });
  await expect(toastLocator.first()).toBeVisible({ timeout });
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
      timeout: options?.timeout ?? 5000,
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

function buildTextMatcher(text: string, exact: boolean): string | RegExp {
  const normalized = text.trim();
  if (!normalized) {
    throw new Error("Toast text must be a non-empty string");
  }

  return exact ? normalized : new RegExp(escapeRegExp(normalized), "i");
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
