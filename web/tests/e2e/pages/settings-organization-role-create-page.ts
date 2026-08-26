import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { RoleCreateFormSection } from "../sections";

import { settingsOrganizationPath } from "@/lib/paths";

/**
 * Page Object Model for Organization Role Create page.
 * Represents the role creation view at /settings/organizations/:organizationSlug/roles/new
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
   * @param organizationSlug - Canonical organization slug.
   */
  async goto(organizationSlug: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `${settingsOrganizationPath({ organizationSlug })}/roles/new`
    );
  }
}
