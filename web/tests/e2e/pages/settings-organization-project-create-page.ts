import type { Page } from "@playwright/test";
import { settingsProjectNewPath } from "@/lib/paths";
import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { ProjectCreateFormSection } from "../sections";

export class SettingsOrganizationProjectCreatePage extends BaseComponent {
  public readonly projectForm: ProjectCreateFormSection;

  constructor(page: Page) {
    super(page);
    this.projectForm = new ProjectCreateFormSection(page);
  }

  async goto(organizationSlug: string, namespaceSlug: string): Promise<void> {
    await navigateAndWait(
      this.page,
      settingsProjectNewPath({ organizationSlug, namespaceSlug })
    );
  }
}
