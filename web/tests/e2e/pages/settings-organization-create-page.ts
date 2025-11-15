import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { OrganizationCreateFormSection } from "../sections";

/**
 * Page Object Model for Organization Create page.
 * Represents the organization create view at /settings/organizations/create
 */
export class SettingsOrganizationCreatePage extends BaseComponent {
  public readonly organizationCreateForm: OrganizationCreateFormSection;

  constructor(page: Page) {
    super(page);
    this.organizationCreateForm = new OrganizationCreateFormSection(page);
  }

  /**
   * Navigate to the organization create page.
   */
  async goto(): Promise<void> {
    await navigateAndWait(this.page, `/settings/organizations/new`);
  }
}
