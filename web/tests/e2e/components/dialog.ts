import { expect } from "@playwright/test";
import type { Page } from "@playwright/test";

import { LocatorComponent } from "./base";
import { closeDialog, confirmDialog, waitForDialog } from "../helpers/dialogs";

/**
 * Reusable Dialog component helper.
 * Provides common dialog operations with a composable API.
 */
export class Dialog extends LocatorComponent {
  constructor(page: Page) {
    const dialogLocator = page
      .getByRole("dialog")
      .or(page.getByRole("alertdialog"))
      .first();
    super(page, dialogLocator);
  }

  /**
   * Wait for dialog to appear.
   */
  async waitFor(title?: string): Promise<void> {
    await waitForDialog(this.page, title);
    // Update locator after dialog appears
    this.locator = this.page
      .getByRole("dialog")
      .or(this.page.getByRole("alertdialog"))
      .first();
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
   * Wait for dialog to disappear.
   */
  async waitForClose(): Promise<void> {
    await expect(this.locator).not.toBeVisible();
  }

  /**
   * Get a button within the dialog.
   * Useful for composing with other components.
   */
  getButton(name: string) {
    return this.locator.getByRole("button", { name });
  }

  /**
   * Get the dialog content locator.
   */
  getContent() {
    return this.locator;
  }
}
