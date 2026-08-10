import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { ProjectCreateFormSection } from "../sections";

export class SettingsOrganizationProjectCreatePage extends BaseComponent {
  public readonly projectForm: ProjectCreateFormSection;

  constructor(page: Page) {
    super(page);
    this.projectForm = new ProjectCreateFormSection(page);
  }

  async goto(organizationId: string, namespaceId: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${organizationId}/namespaces/${namespaceId}/projects/new`
    );
  }
}
