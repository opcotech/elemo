import type { Page } from "@playwright/test";

import { Form } from "../components";
import { SectionContainerMixin } from "../mixins";

/**
 * Reusable Organization Edit Form Section component.
 * Handles organization edit form interactions.
 * Can be composed into any page that displays organization edit form.
 */
export class OrganizationEditFormSection extends SectionContainerMixin(Form) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='organization-edit-form']")
    );
  }

  /**
   * Wait for section to load and be visible.
   */
  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }
}
