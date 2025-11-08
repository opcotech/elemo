import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

/**
 * Find a table by its header text.
 * @param page - Playwright page object
 * @param headerText - Text that appears in one of the table headers
 * @returns Locator for the table
 */
export function getTableByHeader(page: Page, headerText: string): Locator {
  return page
    .locator("table")
    .filter({
      has: page.locator("thead th", { hasText: headerText }),
    })
    .first();
}

/**
 * Find a row in a table that contains the specified text.
 * @param table - Table locator
 * @param text - Text to search for in the row
 * @returns Locator for the table row
 */
export function getTableRow(table: Locator, text: string): Locator {
  return table.locator("tbody tr").filter({ hasText: text }).first();
}

/**
 * Find a button by its accessible name (text content or aria-label).
 * @param page - Playwright page object
 * @param text - Button text or accessible name
 * @param options - Optional options for matching (exact, case-insensitive)
 * @returns Locator for the button
 */
export function getButtonByText(
  page: Page,
  text: string,
  options?: { exact?: boolean; caseInsensitive?: boolean }
): Locator {
  if (options?.caseInsensitive) {
    return page.getByRole("button", { name: new RegExp(text, "i") });
  }
  return page.getByRole("button", { name: text, exact: options?.exact });
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
 * Find an element by text with optional role filter.
 * @param page - Playwright page object
 * @param text - Text to search for
 * @param role - Optional ARIA role to filter by
 * @returns Locator for the element
 */
export function getElementByText(
  page: Page,
  text: string,
  role?: string
): Locator {
  if (role) {
    return page.getByRole(role as any, { name: text });
  }
  return page.getByText(text);
}

/**
 * Wait for a dropdown/combobox to open and be ready for interaction.
 * @param combobox - The combobox locator
 */
export async function waitForDropdownOpen(combobox: Locator): Promise<void> {
  await expect(combobox)
    .toHaveAttribute("aria-expanded", "true", {
      timeout: 5000,
    })
    .catch(() => {});
  await combobox
    .page()
    .waitForFunction(
      () => {
        const dropdowns = document.querySelectorAll(
          '[role="listbox"], [role="menu"]'
        );
        return Array.from(dropdowns).some(
          (d) => window.getComputedStyle(d).display !== "none"
        );
      },
      { timeout: 5000 }
    )
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

  try {
    await skeleton.first().waitFor({
      state: "hidden",
      timeout: options?.timeout ?? 5000,
    });
  } catch {
    // ignore timeout to keep flow resilient
  }
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
 * Wait for a table to be visible when it exists.
 * @param table - Table locator
 * @param options - Optional timeout override
 */
export async function waitForTableReady(
  table: Locator,
  options?: { timeout?: number }
): Promise<void> {
  let count = 0;
  try {
    count = await table.count();
  } catch {
    count = 0;
  }

  if (count === 0) {
    return;
  }

  await waitForElementVisible(table, options);
}

/**
 * Find a button within a container by accessible text or aria-label.
 * @param container - Locator that scopes the search
 * @param name - Accessible name to match
 * @param options - Matching options (exact match)
 * @returns Locator for the first matching button
 */
export function getButtonIn(
  container: Locator,
  name: string,
  options?: { exact?: boolean }
): Locator {
  const hasText = options?.exact ? name : new RegExp(name, "i");
  const textMatch = container.locator("button").filter({ hasText });
  const ariaMatch = container.locator(`button[aria-label="${name}"]`);
  return textMatch.or(ariaMatch).first();
}
