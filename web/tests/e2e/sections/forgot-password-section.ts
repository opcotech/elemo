import type { Page } from "@playwright/test";

import { Form } from "../components";
import { BaseComponent } from "../components/base";
import { waitForElementVisible } from "../helpers";

/**
 * Forgot-password request form section.
 */
export class ForgotPasswordSection extends BaseComponent {
  private form: Form;

  constructor(page: Page) {
    super(page);
    this.form = new Form(page);
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    // CardTitle renders as a div, not a heading role
    await waitForElementVisible(
      this.page.getByText("Forgot your password?", { exact: true }),
      options
    );
  }

  async fillEmail(email: string): Promise<void> {
    await this.form.fillFields({ Email: email });
  }

  async submit(): Promise<void> {
    await this.page.getByRole("button", { name: "Reset password" }).click();
  }

  async requestReset(email: string): Promise<void> {
    await this.fillEmail(email);
    await this.submit();
  }
}
