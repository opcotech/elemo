import { expect } from "@playwright/test";
import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { waitForElementVisible } from "../helpers";

/**
 * Forgot-password request form section.
 */
export class ForgotPasswordSection extends BaseComponent {
  constructor(page: Page) {
    super(page);
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    // CardTitle renders as a div, not a heading role
    await waitForElementVisible(
      this.page.getByText("Forgot your password?", { exact: true }),
      options
    );
  }

  private emailField() {
    return this.page.getByLabel("Email", { exact: true });
  }

  async fillEmail(email: string): Promise<void> {
    const field = this.emailField();
    await expect(async () => {
      await expect(field).toBeEditable();
      await field.evaluate((el, value) => {
        const input = el as HTMLInputElement;
        input.focus();
        input.value = value;
        input.dispatchEvent(new Event("input", { bubbles: true }));
        input.dispatchEvent(new Event("change", { bubbles: true }));
      }, email);
      await expect(field).toHaveValue(email);
    }).toPass({ timeout: 10_000 });
  }

  async submit(): Promise<void> {
    await this.page.getByRole("button", { name: "Reset password" }).click();
  }

  async requestReset(email: string): Promise<void> {
    const success = this.page.getByText(/If an account with this email exists/);

    await expect(async () => {
      await this.fillEmail(email);
      await this.submit();
      await expect(success).toBeVisible({ timeout: 4_000 });
    }).toPass({ timeout: 20_000 });
  }
}
