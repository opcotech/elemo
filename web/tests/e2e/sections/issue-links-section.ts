import type { Locator, Page } from "@playwright/test";

import { Form } from "../components";
import { clickUntilVisible, fillLocator } from "../helpers";
import { DialogMixin, SectionContainerMixin } from "../mixins";

/**
 * Issue links section: add, edit, and remove external URLs.
 */
export class IssueLinksSection extends DialogMixin(
  SectionContainerMixin(Form)
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(this.page.locator("[data-section='issue-links']"));
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  getAddButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Add",
      exact: true,
    });
  }

  async clickAdd(): Promise<void> {
    await clickUntilVisible(
      this.getAddButton(),
      this.getDialog().getContent()
    );
    await this.waitForDialog("Add link");
  }

  getLink(displayLabel: string): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: `Edit link ${displayLabel}`,
    });
  }

  getRemoveLinkButton(displayLabel: string): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: `Remove link ${displayLabel}`,
    });
  }

  async fillLinkFields(fields: { url: string; label?: string }): Promise<void> {
    const dialog = this.getDialog().getContent();
    await fillLocator(dialog.getByLabel("URL"), fields.url);
    if (fields.label !== undefined) {
      await fillLocator(dialog.getByLabel("Label"), fields.label);
    }
  }

  async addLink(fields: { url: string; label?: string }): Promise<void> {
    await this.clickAdd();
    await this.fillLinkFields(fields);
    await this.confirmDialog("Add link");
  }

  async editLink(
    displayLabel: string,
    fields: { url: string; label?: string }
  ): Promise<void> {
    await this.getLink(displayLabel).click();
    await this.waitForDialog("Edit link");
    await this.fillLinkFields(fields);
    await this.confirmDialog("Save");
  }

  async removeLink(displayLabel: string): Promise<void> {
    await this.getRemoveLinkButton(displayLabel).click();
  }

  async removeLinkFromDialog(displayLabel: string): Promise<void> {
    await this.getLink(displayLabel).click();
    await this.waitForDialog("Edit link");
    await this.getDialog().getButton("Remove link").click();
  }
}
