import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { SectionContainerMixin } from "../mixins";

/**
 * Project Danger Zone Section for project deletion.
 */
export class ProjectDangerZoneSection extends SectionContainerMixin(
  BaseComponent
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='project-danger-zone']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  async isVisible(): Promise<boolean> {
    return await this.getSectionContainer()
      .isVisible({ timeout: 2000 })
      .catch(() => false);
  }

  getDeleteButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Delete Project",
    });
  }

  async hasDeleteButton(): Promise<boolean> {
    return await this.getDeleteButton()
      .isVisible({ timeout: 2000 })
      .catch(() => false);
  }

  async clickDeleteButton(): Promise<void> {
    await this.getDeleteButton().click();
  }
}
