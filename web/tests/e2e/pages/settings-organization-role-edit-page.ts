import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { RoleEditFormSection } from "../sections";

import { settingsOrganizationPath } from "@/lib/paths";

/**
 * Page Object Model for Organization Role Edit page.
 * Represents the role edit view at /settings/organizations/:organizationSlug/roles/:roleId/edit
 */
export class SettingsOrganizationRoleEditPage extends BaseComponent {
  public readonly roleEditForm: RoleEditFormSection;

  constructor(page: Page) {
    super(page);
    this.roleEditForm = new RoleEditFormSection(page);
  }

  /**
   * Navigate to the role edit page.
   *
   * @param organizationSlug - Canonical organization slug.
   * @param roleId - The ID of the role to edit.
   */
  async goto(organizationSlug: string, roleId: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `${settingsOrganizationPath({ organizationSlug })}/roles/${roleId}/edit`
    );
  }
}
