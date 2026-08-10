import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { ResetPasswordSection } from "../sections/reset-password-section";

export class ResetPasswordPage extends BaseComponent {
  public readonly form: ResetPasswordSection;

  constructor(page: Page) {
    super(page);
    this.form = new ResetPasswordSection(page);
  }

  async goto(token: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `/reset-password?token=${encodeURIComponent(token)}`
    );
  }
}
