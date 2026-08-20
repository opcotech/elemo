import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { RoleCreateFormSection } from "../sections";

/**
 * Page Object Model for Organization Role Create page.
 * Represents the role creation view at /settings/organizations/:id/roles/new
 */
export class SettingsOrganizationRoleCreatePage extends BaseComponent {
  public readonly roleCreateForm: RoleCreateFormSection;

  constructor(page: Page) {
    super(page);
    this.roleCreateForm = new RoleCreateFormSection(page);
  }

  /**
   * Navigate to the role create page.
   *
   * @param organizationId - The ID of the organization to create role in.
   */
  async goto(organizationId: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${organizationId}/roles/new`
    );
  }
}
