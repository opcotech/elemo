import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { ForgotPasswordSection } from "../sections/forgot-password-section";

export class ForgotPasswordPage extends BaseComponent {
  public readonly form: ForgotPasswordSection;

  constructor(page: Page) {
    super(page);
    this.form = new ForgotPasswordSection(page);
  }

  async goto(): Promise<void> {
    await navigateAndWait(this.page, "/forgot-password");
  }
}
