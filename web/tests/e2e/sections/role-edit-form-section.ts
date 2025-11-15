import type { Page } from "@playwright/test";

import { Form } from "../components";
import { SectionContainerMixin } from "../mixins";

/**
 * Reusable Role Edit Form Section component.
 * Handles role edit form interactions.
 * Can be composed into any page that displays role edit form.
 */
export class RoleEditFormSection extends SectionContainerMixin(Form) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='role-edit-form']")
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
