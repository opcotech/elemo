import type { Page } from "@playwright/test";

import { Form } from "../components";
import { SectionContainerMixin } from "../mixins";

/**
 * Reusable Role Create Form Section component.
 * Handles role create form interactions.
 * Can be composed into any page that displays role create form.
 */
export class RoleCreateFormSection extends SectionContainerMixin(Form) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='role-create-form']")
    );
  }

  /**
   * Wait for section to load and be visible.
   */
  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  /**
   * Click cancel button to return to organization details.
   */
  async cancel(): Promise<void> {
    const cancelButton = this.page.getByRole("button", { name: "Cancel" });
    await cancelButton.click();
  }
}
