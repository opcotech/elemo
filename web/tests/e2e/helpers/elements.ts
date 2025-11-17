import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

import { waitForPageLoad } from "./navigation";

/**
 * Find an element by text with optional role filter.
 * @param locator - Locator to search within
 * @param text - Text to search for
 * @param role - Optional ARIA role to filter by
 * @param options - Optional options for matching (exact, case-insensitive)
 * @returns Locator for the element
 */
export function getElementByText(
  locator: Locator,
  text: string,
  options?: {
    role?: string;
    exact?: boolean;
    caseInsensitive?: boolean;
    nth?: number;
  }
): Locator {
  const { role, exact, caseInsensitive, nth } = options ?? {
    role: undefined,
    exact: true,
    caseInsensitive: false,
    nth: 0,
  };

  if (role) {
    if (caseInsensitive) {
      return locator
        .getByRole(role as any, { name: new RegExp(text, "i") })
        .nth(nth ?? 0);
    }
    return locator.getByRole(role as any, { name: text, exact }).nth(nth ?? 0);
  }

  if (caseInsensitive) {
    return locator.getByText(new RegExp(text, "i")).nth(nth ?? 0);
  }

  return locator.getByText(text, { exact }).nth(nth ?? 0);
}

/**
 * Wait for an element to be visible with proper retry logic.
 * This uses Playwright's built-in auto-waiting, so it's more reliable than manual waits.
 * @param locator - Element locator
 * @param options - Optional timeout and other options
 */
export async function waitForElementVisible(
  locator: Locator,
  options?: { timeout?: number }
): Promise<void> {
  await expect(locator).toBeVisible({
    timeout: options?.timeout ?? 10000,
  });
}

/**
 * Wait for a dropdown/combobox to open and be ready for interaction.
 * @param combobox - The combobox locator
 */
export async function waitForDropdownOpen(combobox: Locator): Promise<void> {
  await expect(combobox)
    .toHaveAttribute("aria-expanded", "true")
    .catch(() => {});
  await combobox
    .page()
    .waitForFunction(() => {
      const dropdowns = document.querySelectorAll(
        '[role="listbox"], [role="menu"]'
      );
      return Array.from(dropdowns).some(
        (d) => window.getComputedStyle(d).display !== "none"
      );
    })
    .catch(() => {});
}

/**
 * Wait for skeleton loaders within a container to disappear, if present.
 * @param container - Locator that may contain skeleton elements
 * @param options - Optional selector and timeout overrides
 */
export async function waitForSkeletonToDisappear(
  container: Locator,
  options?: { selector?: string; timeout?: number }
): Promise<void> {
  const selector = options?.selector ?? '[data-slot="skeleton"]';
  const skeleton = container.locator(selector);
  let hasSkeleton = false;

  try {
    hasSkeleton = (await skeleton.count()) > 0;
  } catch {
    hasSkeleton = false;
  }

  if (!hasSkeleton) {
    return;
  }

  await skeleton
    .first()
    .waitFor({
      state: "hidden",
      timeout: options?.timeout ?? 5000,
    })
    .catch(() => {});
}

/**
 * Wait for the first element in a locator collection to attach, if any.
 * @param locator - Locator collection to monitor
 * @param options - Optional timeout override
 */
export async function waitForLocatorAttached(
  locator: Locator,
  options?: { timeout?: number }
): Promise<void> {
  try {
    await locator.first().waitFor({
      state: "attached",
      timeout: options?.timeout ?? 5000,
    });
  } catch {
    // swallow attachment wait failures
  }
}

/**
 * Wait for URL to be loaded.
 *
 * @param page - Playwright page object
 * @param url - URL to wait for
 */
export async function waitForUrlLoaded(
  page: Page,
  url: string | RegExp,
  options?: { timeout?: number }
): Promise<void> {
  await page.waitForURL(url, {
    timeout: options?.timeout,
  });
  await waitForPageLoad(page);
}
