import type { Page } from "@playwright/test";

import { Form } from "../components";
import { SectionContainerMixin } from "../mixins";

/**
 * Reusable Project Edit Form Section component.
 */
export class ProjectEditFormSection extends SectionContainerMixin(Form) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='project-edit-form']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  async selectStatus(status: "Active" | "Pending"): Promise<void> {
    const container = this.getSectionContainer();
    const trigger = container
      .locator("[data-slot='field']")
      .filter({ has: this.page.getByText("Status", { exact: true }) })
      .getByRole("combobox");
    await trigger.click();
    await this.page.getByRole("option", { name: status, exact: true }).click();
  }

  async cancel(): Promise<void> {
    const cancelButton = this.page.getByRole("button", { name: "Cancel" });
    await cancelButton.click();
  }
}
