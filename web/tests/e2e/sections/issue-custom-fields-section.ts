import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { fillLocator, waitForElementVisible } from "../helpers";
import { SectionContainerMixin } from "../mixins";

export class IssueCustomFieldsSection extends SectionContainerMixin(
  BaseComponent
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='issue-custom-fields']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
    await waitForElementVisible(
      this.getSectionContainer().locator("[data-custom-field-key]").first(),
      options
    );
  }

  getFieldInput(name: string): Locator {
    return this.getSectionContainer().getByRole("textbox", {
      name,
      exact: true,
    });
  }

  async fillTextField(name: string, value: string): Promise<void> {
    await fillLocator(this.getFieldInput(name), value);
    await this.getFieldInput(name).blur();
  }
}
