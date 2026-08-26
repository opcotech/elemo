import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait, waitForElementVisible } from "../helpers";
import { ProjectDangerZoneSection, ProjectInfoSection } from "../sections";

import { settingsProjectPath } from "@/lib/paths";

export class SettingsOrganizationProjectDetailsPage extends BaseComponent {
  public readonly projectInfo: ProjectInfoSection;
  public readonly dangerZone: ProjectDangerZoneSection;

  constructor(page: Page) {
    super(page);
    this.projectInfo = new ProjectInfoSection(page);
    this.dangerZone = new ProjectDangerZoneSection(page);
  }

  async goto(
    organizationSlug: string,
    namespaceSlug: string,
    projectKey: string
  ): Promise<void> {
    await navigateAndWait(
      this.page,
      settingsProjectPath({ organizationSlug, namespaceSlug, projectKey })
    );
  }

  async waitForLoad(): Promise<void> {
    // Prefer settled project content over the first transient h1 (e.g. still
    // on "Create Project" while .../projects/new matches a loose URL assert).
    const settled = this.page
      .getByText("Project Information")
      .or(this.page.getByRole("heading", { name: "Access Denied" }));
    await waitForElementVisible(settled);
  }

  async getTitleText(): Promise<string> {
    await this.waitForLoad();
    const heading = this.page
      .getByRole("main")
      .getByRole("heading", { level: 1 })
      .filter({ hasNotText: "Create Project" })
      .first();
    await waitForElementVisible(heading);
    return (await heading.textContent()) ?? "";
  }
}
