import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { OrganizationEditFormSection } from "../sections";

/**
 * Page Object Model for Organization Edit page.
 * Represents the organization edit view at /settings/organizations/:id/edit
 */
export class SettingsOrganizationEditPage extends BaseComponent {
  public readonly organizationEditForm: OrganizationEditFormSection;

  constructor(page: Page) {
    super(page);
    this.organizationEditForm = new OrganizationEditFormSection(page);
  }

  /**
   * Navigate to the organization edit page.
   *
   * @param organizationId - The ID of the organization to navigate to.
   */
  async goto(organizationId: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${organizationId}/edit`
    );
  }
}
