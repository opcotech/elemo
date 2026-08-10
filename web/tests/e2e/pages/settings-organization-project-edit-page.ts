import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait } from "../helpers";
import { ProjectEditFormSection } from "../sections";

export class SettingsOrganizationProjectEditPage extends BaseComponent {
  public readonly projectEditForm: ProjectEditFormSection;

  constructor(page: Page) {
    super(page);
    this.projectEditForm = new ProjectEditFormSection(page);
  }

  async goto(
    organizationId: string,
    namespaceId: string,
    projectId: string
  ): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${organizationId}/namespaces/${namespaceId}/projects/${projectId}/edit`
    );
  }
}
