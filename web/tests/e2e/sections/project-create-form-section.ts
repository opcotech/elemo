import type { Page } from "@playwright/test";

import { Form } from "../components";
import { SectionContainerMixin } from "../mixins";

/**
 * Reusable Project Create Form Section component.
 */
export class ProjectCreateFormSection extends SectionContainerMixin(Form) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='project-create-form']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  async cancel(): Promise<void> {
    const cancelButton = this.page.getByRole("button", { name: "Cancel" });
    await cancelButton.click();
  }
}
