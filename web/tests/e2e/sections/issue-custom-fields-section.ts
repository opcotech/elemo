import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { fillLocator } from "../helpers";
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

  getFieldInput(name: string): Locator {
    return this.getSectionContainer().getByLabel(name, { exact: true });
  }

  async fillTextField(name: string, value: string): Promise<void> {
    await fillLocator(this.getFieldInput(name), value);
    await this.getFieldInput(name).blur();
  }
}
