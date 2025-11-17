import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { LoginSection } from "../sections/login-section";

/**
 * Page Object Model for Login page.
 * Composed from reusable LoginSection.
 */
export class LoginPage extends BaseComponent {
  public readonly login: LoginSection;

  constructor(page: Page) {
    super(page);
    this.login = new LoginSection(page);
  }

  /**
   * Navigate to the login page.
   */
  async goto(): Promise<void> {
    await navigateAndWait(this.page, "/login");
  }
}
