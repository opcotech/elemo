import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { ProjectEditFormSection } from "../sections";

import { settingsProjectEditPath } from "@/lib/paths";

export class SettingsOrganizationProjectEditPage extends BaseComponent {
  public readonly projectEditForm: ProjectEditFormSection;

  constructor(page: Page) {
    super(page);
    this.projectEditForm = new ProjectEditFormSection(page);
  }

  async goto(
    organizationSlug: string,
    namespaceSlug: string,
    projectKey: string
  ): Promise<void> {
    await navigateAndWait(
      this.page,
      settingsProjectEditPath({ organizationSlug, namespaceSlug, projectKey })
    );
  }
}
