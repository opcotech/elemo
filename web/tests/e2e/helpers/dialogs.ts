import { expect } from "@playwright/test";
import type { Page } from "@playwright/test";

/**
 * Wait for a dialog to appear.
 * @param page - Playwright page object
 * @param title - Optional dialog title to wait for
 * @param options - Optional timeout
 */
export async function waitForDialog(
  page: Page,
  title?: string,
  options?: { timeout?: number }
): Promise<void> {
  const timeout = options?.timeout ?? 5000;

  if (title) {
    const dialogLocator = page
      .getByRole("dialog")
      .or(page.getByRole("alertdialog"));
    await expect(dialogLocator).toBeVisible({ timeout });
    await expect(
      dialogLocator.getByRole("heading", { name: title })
    ).toBeVisible({ timeout });
  } else {
    const dialogLocator = page
      .getByRole("dialog")
      .or(page.getByRole("alertdialog"));
    await expect(dialogLocator).toBeVisible({ timeout });
  }
}

/**
 * Close a dialog by clicking cancel or close button.
 * @param page - Playwright page object
 * @param action - Action to take: "cancel" or "close"
 */
export async function closeDialog(
  page: Page,
  action: "cancel" | "close" = "cancel"
): Promise<void> {
  const dialog = page
    .getByRole("dialog")
    .or(page.getByRole("alertdialog"))
    .first();

  if (action === "cancel") {
    const cancelButton = dialog.getByRole("button", { name: "Cancel" });
    await cancelButton.click();
  } else {
    const closeButton = dialog.getByRole("button", { name: /close/i });
    await closeButton.click();
  }
  await expect(dialog).not.toBeVisible();
}

/**
 * Confirm a dialog action by clicking the confirm/delete button.
 * @param page - Playwright page object
 * @param buttonText - Optional specific button text (defaults to "Delete" or "Confirm")
 */
export async function confirmDialog(
  page: Page,
  buttonText?: string
): Promise<void> {
  const dialog = page
    .getByRole("dialog")
    .or(page.getByRole("alertdialog"))
    .first();

  if (buttonText) {
    const confirmButton = dialog.getByRole("button", { name: buttonText });
    await confirmButton.click();
  } else {
    const confirmButton = dialog.getByRole("button", {
      name: /^(Delete|Confirm|Save|Submit|OK)$/i,
    });
    await confirmButton.click();
  }
  await expect(dialog).not.toBeVisible();
}
