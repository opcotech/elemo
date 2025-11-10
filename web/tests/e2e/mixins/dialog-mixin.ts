import type { Page } from "@playwright/test";

import { Dialog } from "../components";

/**
 * Mixin for sections that interact with dialogs.
 * Provides common dialog interaction methods.
 */
export function DialogMixin<T extends abstract new (...args: any[]) => any>(
  Base: T
) {
  abstract class DialogMixinClass extends Base {
    protected dialog: Dialog;

    constructor(...args: any[]) {
      super(...args);
      // Extract page from constructor args (assumes first arg is page for BaseComponent)
      const page = args[0] as Page;
      this.dialog = new Dialog(page);
    }

    /**
     * Get the dialog instance for direct access.
     */
    protected getDialog(): Dialog {
      return this.dialog;
    }

    /**
     * Wait for a dialog to appear with optional title.
     */
    protected async waitForDialog(title?: string): Promise<void> {
      await this.dialog.waitFor(title);
    }

    /**
     * Confirm the dialog action.
     */
    protected async confirmDialog(buttonText?: string): Promise<void> {
      await this.dialog.confirm(buttonText);
    }

    /**
     * Cancel/close the dialog.
     */
    protected async cancelDialog(): Promise<void> {
      await this.dialog.cancel();
    }

    /**
     * Wait for dialog to close.
     */
    protected async waitForDialogClose(): Promise<void> {
      await this.dialog.waitForClose();
    }
  }

  return DialogMixinClass;
}
