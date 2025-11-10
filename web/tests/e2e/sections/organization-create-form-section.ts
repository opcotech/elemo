import type { Page } from "@playwright/test";

import { Form } from "../components";
import { SectionContainerMixin } from "../mixins";

/**
 * Reusable Organization Create Form Section component.
 * Handles organization create form interactions.
 * Can be composed into any page that displays organization create form.
 */
export class OrganizationCreateFormSection extends SectionContainerMixin(Form) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='organization-create-form']")
    );
  }

  /**
   * Wait for section to load and be visible.
   */
  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }
}
