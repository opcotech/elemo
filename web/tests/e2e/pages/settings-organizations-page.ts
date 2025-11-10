import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { OrganizationsSection } from "../sections/organizations-section";

/**
 * Page Object Model for Settings Organizations page.
 * Composed from reusable OrganizationsSection.
 */
export class SettingsOrganizationsPage extends BaseComponent {
  public readonly organizations: OrganizationsSection;

  constructor(page: Page) {
    super(page);
    this.organizations = new OrganizationsSection(page);
  }

  /**
   * Navigate to the settings organizations page.
   */
  async goto(): Promise<void> {
    await navigateAndWait(this.page, "/settings/organizations");
  }
}
