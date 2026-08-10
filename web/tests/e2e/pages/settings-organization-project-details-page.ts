import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait, waitForElementVisible } from "../helpers";
import {
  ProjectDangerZoneSection,
  ProjectDocumentsSection,
  ProjectInfoSection,
  ProjectIssuesSection,
} from "../sections";

export class SettingsOrganizationProjectDetailsPage extends BaseComponent {
  public readonly projectInfo: ProjectInfoSection;
  public readonly documents: ProjectDocumentsSection;
  public readonly issues: ProjectIssuesSection;
  public readonly dangerZone: ProjectDangerZoneSection;

  constructor(page: Page) {
    super(page);
    this.projectInfo = new ProjectInfoSection(page);
    this.documents = new ProjectDocumentsSection(page);
    this.issues = new ProjectIssuesSection(page);
    this.dangerZone = new ProjectDangerZoneSection(page);
  }

  async goto(
    organizationId: string,
    namespaceId: string,
    projectId: string
  ): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${organizationId}/namespaces/${namespaceId}/projects/${projectId}`
    );
  }

  async waitForLoad(): Promise<void> {
    const heading = this.page
      .getByRole("main")
      .getByRole("heading", { level: 1 })
      .first();
    await waitForElementVisible(heading);
  }

  async getTitleText(): Promise<string> {
    const heading = this.page
      .getByRole("main")
      .getByRole("heading", { level: 1 })
      .first();
    await waitForElementVisible(heading);
    return (await heading.textContent()) ?? "";
  }
}
