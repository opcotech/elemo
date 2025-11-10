import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { SectionContainerMixin } from "../mixins";

/**
 * Reusable Organization Danger Zone Section component.
 * Handles danger zone interactions for organization deletion.
 * Can be composed into any page that displays the danger zone.
 */
export class OrganizationDangerZoneSection extends SectionContainerMixin(
  BaseComponent
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='organization-danger-zone']")
    );
  }

  /**
   * Wait for danger zone section to load and be visible.
   */
  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  /**
   * Returns whether the danger zone section is visible.
   */
  async isVisible(): Promise<boolean> {
    return await this.getSectionContainer().isVisible({ timeout: 2000 });
  }

  /**
   * Get the delete organization button locator.
   */
  getDeleteButton(): Locator {
    return this.page.getByRole("button", { name: "Delete Organization" });
  }

  /**
   * Check if the delete organization button is visible.
   */
  async hasDeleteButton(): Promise<boolean> {
    const button = this.getDeleteButton();
    return await button.isVisible({ timeout: 2000 }).catch(() => false);
  }

  /**
   * Click on the delete organization button.
   */
  async clickDeleteButton(): Promise<void> {
    await this.getDeleteButton().click();
  }
}
