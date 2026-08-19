import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { clickUntilURL } from "../helpers";
import { SectionContainerMixin } from "../mixins";

/**
 * Reusable Organization Info Section component.
 * Handles organization info interactions.
 * Can be composed into any page that displays organization info.
 */
export class OrganizationInfoSection extends SectionContainerMixin(
  BaseComponent
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='organization-info']")
    );
  }

  /**
   * Wait for section to load and be visible.
   */
  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  /**
   * Get the edit organization button locator.
   */
  getEditOrganizationButton(): Locator {
    return this.getSectionContainer().getByRole("link", { name: /edit/i });
  }

  /**
   * Check if the edit organization button is visible.
   */
  async hasEditOrganizationButton(): Promise<boolean> {
    const button = this.getEditOrganizationButton();
    return await button.isVisible().catch(() => false);
  }

  /**
   * Click on the edit organization button.
   */
  async clickEditOrganizationButton(): Promise<void> {
    await clickUntilURL(
      this.getEditOrganizationButton(),
      /\/settings\/organizations\/[^/]+\/edit$/
    );
  }
}
