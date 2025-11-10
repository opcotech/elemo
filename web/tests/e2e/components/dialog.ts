import { expect } from "@playwright/test";
import { closeDialog, confirmDialog, waitForDialog } from "../helpers/dialogs";
import type { Page } from "@playwright/test";

/**
 * Reusable Dialog component helper.
 * Provides common dialog operations.
 */
export class Dialog {
  constructor(protected page: Page) {}

  /**
   * Wait for dialog to appear.
   */
  async waitFor(title?: string): Promise<void> {
    await waitForDialog(this.page, title);
  }

  /**
   * Close the dialog by clicking cancel.
   */
  async cancel(): Promise<void> {
    await closeDialog(this.page, "cancel");
  }

  /**
   * Confirm the dialog action.
   */
  async confirm(buttonText?: string): Promise<void> {
    await confirmDialog(this.page, buttonText);
  }

  /**
   * Get the dialog locator.
   */
  getLocator() {
    return this.page
      .getByRole("dialog")
      .or(this.page.getByRole("alertdialog"))
      .first();
  }

  /**
   * Check if dialog is visible.
   */
  async isVisible(): Promise<boolean> {
    return await this.getLocator()
      .isVisible()
      .catch(() => false);
  }

  /**
   * Wait for dialog to disappear.
   */
  async waitForClose(): Promise<void> {
    await expect(this.getLocator()).not.toBeVisible();
  }
}
