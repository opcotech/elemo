import type { Page } from "@playwright/test";

import { Form } from "../components";
import { BaseComponent } from "../components/base";
import { waitForElementVisible } from "../helpers";

/**
 * Reset-password form section (token URL).
 */
export class ResetPasswordSection extends BaseComponent {
  private form: Form;

  constructor(page: Page) {
    super(page);
    this.form = new Form(page);
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    // CardTitle renders as a div, not a heading role
    await waitForElementVisible(
      this.page.getByText("Reset your password", { exact: true }),
      options
    );
  }

  async fillPasswordFields(fields: {
    password: string;
    confirmPassword: string;
  }): Promise<void> {
    await this.form.fillFields({
      "New Password": fields.password,
      "Confirm New Password": fields.confirmPassword,
    });
  }

  async submit(): Promise<void> {
    await this.page.getByRole("button", { name: "Reset password" }).click();
  }

  async resetPassword(
    password: string,
    confirmPassword?: string
  ): Promise<void> {
    await this.fillPasswordFields({
      password,
      confirmPassword: confirmPassword ?? password,
    });
    await this.submit();
  }

  getNewPasswordField() {
    return this.page.getByLabel("New Password", { exact: true });
  }

  getConfirmPasswordField() {
    return this.page.getByLabel("Confirm New Password", { exact: true });
  }
}
