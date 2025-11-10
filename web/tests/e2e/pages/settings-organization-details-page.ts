import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import {
  NamespacesSection,
  OrganizationDangerZoneSection,
  OrganizationInfoSection,
  OrganizationMembersSection,
  RolesSection,
} from "../sections";

/**
 * Page Object Model for Organization Details page.
 * Represents the organization detail view at /settings/organizations/:id
 */
export class SettingsOrganizationDetailsPage extends BaseComponent {
  public readonly organizationInfo: OrganizationInfoSection;
  public readonly dangerZone: OrganizationDangerZoneSection;
  public readonly members: OrganizationMembersSection;
  public readonly namespaces: NamespacesSection;
  public readonly roles: RolesSection;

  constructor(page: Page) {
    super(page);
    this.organizationInfo = new OrganizationInfoSection(page);
    this.dangerZone = new OrganizationDangerZoneSection(page);
    this.members = new OrganizationMembersSection(page);
    this.namespaces = new NamespacesSection(page);
    this.roles = new RolesSection(page);
  }

  /**
   * Navigate to the organization details page.
   *
   * @param organizationId - The ID of the organization to navigate to.
   */
  async goto(organizationId: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${organizationId}`
    );
  }
}
