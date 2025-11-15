import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import {
  RoleEditFormSection,
  RoleMembersSection,
  RolePermissionsSection,
} from "../sections";

/**
 * Page Object Model for Organization Role Edit page.
 * Represents the role edit view at /settings/organizations/:id/roles/:roleId/edit
 */
export class SettingsOrganizationRoleEditPage extends BaseComponent {
  public readonly roleEditForm: RoleEditFormSection;
  public readonly members: RoleMembersSection;
  public readonly permissions: RolePermissionsSection;

  constructor(page: Page) {
    super(page);
    this.roleEditForm = new RoleEditFormSection(page);
    this.members = new RoleMembersSection(page);
    this.permissions = new RolePermissionsSection(page);
  }

  /**
   * Navigate to the role edit page.
   *
   * @param organizationId - The ID of the organization.
   * @param roleId - The ID of the role to edit.
   */
  async goto(organizationId: string, roleId: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${organizationId}/roles/${roleId}/edit`
    );
  }
}
