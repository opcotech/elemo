import type { Page } from "@playwright/test";

import { waitForElementVisible } from "./elements";

/**
 * Wait for a card to be visible.
 *
 * @param card - The card locator
 * @param options - Optional timeout and other options
 * @returns Promise<void>
 */
export async function waitForCardVisible(
  page: Page,
  title: string,
  options?: { timeout?: number }
): Promise<void> {
  await waitForElementVisible(
    page.locator("[data-slot='card-header']").getByRole("heading", {
      name: title,
      level: 3,
    }),
    { timeout: options?.timeout ?? 5000 }
  );
}
