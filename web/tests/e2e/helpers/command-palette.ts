import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

import { waitForAnimations } from "./elements";

/**
 * Open the command palette and wait for it to be interactive.
 */
export async function openCommandPalette(page: Page): Promise<Locator> {
  await page.keyboard.press("Control+K");
  const commandDialog = page.getByRole("dialog", {
    name: "Search or run a command",
  });
  await expect(commandDialog).toBeVisible();
  await waitForAnimations(commandDialog);
  return commandDialog;
}

/**
 * Run a command palette action by title.
 */
export async function runPaletteCommand(
  page: Page,
  commandTitle: string
): Promise<void> {
  const commandDialog = await openCommandPalette(page);
  await commandDialog
    .getByPlaceholder("Search entities, navigation, or commands...")
    .fill(commandTitle);
  await commandDialog
    .locator('[data-slot="command-item"]')
    .filter({ hasText: commandTitle })
    .first()
    .click();
  await expect(commandDialog).not.toBeVisible();
}

/**
 * Sequence shortcuts are handled only while the command palette is open.
 */
export async function pressPaletteShortcut(
  page: Page,
  secondKey: "s" | "n"
): Promise<void> {
  await openCommandPalette(page);
  await page.keyboard.down("Shift");
  await page.keyboard.press("t");
  await page.keyboard.press(secondKey);
  await page.keyboard.up("Shift");
}

/**
 * Search from the command palette and return the matching result item.
 */
export async function searchPaletteResult(
  page: Page,
  query: string
): Promise<Locator> {
  const commandDialog = await openCommandPalette(page);
  await commandDialog
    .getByPlaceholder("Search entities, navigation, or commands...")
    .fill(query);
  const result = commandDialog
    .locator('[data-slot="command-item"]')
    .filter({ hasText: query })
    .first();
  await expect(async () => {
    if (!(await result.isVisible().catch(() => false))) {
      await commandDialog
        .getByPlaceholder("Search entities, navigation, or commands...")
        .fill(query);
    }
    await expect(result).toBeVisible({ timeout: 3_000 });
  }).toPass({ timeout: 30_000 });
  return result;
}
