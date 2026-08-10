import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { getElementByText } from "../helpers";
import { SectionContainerMixin } from "../mixins";

/**
 * Reusable Project Info Section component.
 */
export class ProjectInfoSection extends SectionContainerMixin(BaseComponent) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='project-info']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  getFieldValue(label: string): Locator {
    return this.getSectionContainer()
      .locator("label", { hasText: new RegExp(`^${label}$`) })
      .locator("xpath=..")
      .locator(":scope > p, :scope > div")
      .first();
  }

  getEditProjectButton(): Locator {
    return getElementByText(this.getSectionContainer(), "Edit");
  }

  async hasEditProjectButton(): Promise<boolean> {
    const button = this.getEditProjectButton();
    return await button.isVisible().catch(() => false);
  }

  async clickEditProjectButton(): Promise<void> {
    await this.getEditProjectButton().click();
  }
}
