import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";

/**
 * Install an instance plugin, or leave it in place when another spec already
 * installed the same package. The catalog list is client-rendered, so this
 * waits for the empty state or a plugin row before touching the file input.
 * A parallel install can win the race and surface an error toast; a reload
 * then shows the package the other request stored.
 */
export async function ensureInstancePlugin(
  page: Page,
  pluginId: string,
  zipPath: string
): Promise<void> {
  await page.goto("/settings/plugins");

  const row = page.locator("li").filter({ hasText: pluginId });
  // The list query renders a busy skeleton until the client fetch finishes.
  // Sidebar items are already in the document, so they are not a ready signal.
  await expect(page.getByRole("heading", { name: "Plugins" })).toBeVisible();
  await expect(page.locator('[aria-busy="true"]')).toHaveCount(0, {
    timeout: 30_000,
  });

  if ((await row.count()) > 0) {
    return;
  }

  // The install input is mounted with the button. Setting files directly
  // avoids the native file chooser that the button click opens.
  await page.locator('input[type="file"]').first().setInputFiles(zipPath);

  const toast = page.locator("[data-sonner-toast]");
  await expect(toast.or(row).first()).toBeVisible({ timeout: 30_000 });

  if ((await row.count()) === 0) {
    await page.reload();
    await expect(page.getByRole("heading", { name: "Plugins" })).toBeVisible();
    await expect(page.locator('[aria-busy="true"]')).toHaveCount(0, {
      timeout: 30_000,
    });
  }

  await expect(row).toBeVisible({ timeout: 30_000 });
}
