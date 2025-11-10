import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { SecuritySection } from "../sections/security-section";

/**
 * Page Object Model for Settings Security page.
 * Composed from reusable SecuritySection.
 */
export class SettingsSecurityPage extends BaseComponent {
  public readonly security: SecuritySection;

  constructor(page: Page) {
    super(page);
    this.security = new SecuritySection(page);
  }

  /**
   * Navigate to the settings security page.
   */
  async goto(): Promise<void> {
    await navigateAndWait(this.page, "/settings/security");
  }
}
