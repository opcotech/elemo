import type { Page } from "@playwright/test";
import { settingsOrganizationEditPath } from "@/lib/paths";
import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { OrganizationEditFormSection } from "../sections";

/**
 * Page Object Model for Organization Edit page.
 * Represents the organization edit view at /settings/organizations/:organizationSlug/edit
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
   * @param organizationSlug - Canonical organization slug.
   */
  async goto(organizationSlug: string): Promise<void> {
    await navigateAndWait(
      this.page,
      settingsOrganizationEditPath({ organizationSlug })
    );
  }
}
