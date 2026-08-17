import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { clickUntilVisible, fillLocator } from "../helpers";
import { SectionContainerMixin } from "../mixins";

/**
 * Always-on document editor: title, excerpt, body, toolbar, and TOC.
 */
export class DocumentEditorSection extends SectionContainerMixin(
  BaseComponent
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='document-editor']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  getTitleInput(): Locator {
    return this.getSectionContainer().getByLabel("Document title");
  }

  getExcerptInput(): Locator {
    return this.getSectionContainer().getByLabel("Document excerpt");
  }

  getContentEditor(): Locator {
    return this.getSectionContainer().getByLabel("Document content");
  }

  getSaveButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Save",
      exact: true,
    });
  }

  getDiscardButton(): Locator {
    return this.getSectionContainer().getByRole("button", { name: "Discard" });
  }

  getToolbarButton(label: string): Locator {
    return this.getSectionContainer().getByLabel(label, { exact: true });
  }

  getTocToggle(): Locator {
    return this.getSectionContainer().getByLabel(/table of contents/i);
  }

  getToc(): Locator {
    return this.getSectionContainer().getByRole("navigation", {
      name: "Table of contents",
    });
  }

  getTocHeading(name: string): Locator {
    return this.getToc().getByRole("button", { name, exact: true });
  }

  getMoreActionsButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "More actions",
    });
  }

  async editTitle(title: string): Promise<void> {
    await fillLocator(this.getTitleInput(), title);
  }

  async editExcerpt(excerpt: string): Promise<void> {
    await fillLocator(this.getExcerptInput(), excerpt);
  }

  async typeContent(text: string): Promise<void> {
    const editor = this.getContentEditor();
    await editor.click();
    await editor.pressSequentially(text, { delay: 10 });
  }

  async save(): Promise<void> {
    await this.getSaveButton().click();
  }

  async discard(): Promise<void> {
    await this.getDiscardButton().click();
  }

  async openToc(): Promise<void> {
    if (
      await this.getToc()
        .isVisible()
        .catch(() => false)
    ) {
      return;
    }
    await clickUntilVisible(this.getTocToggle(), this.getToc());
  }

  async openMoreActions(): Promise<void> {
    await clickUntilVisible(
      this.getMoreActionsButton(),
      this.page.locator('[data-slot="dropdown-menu-content"]')
    );
  }

  async clickMoreAction(name: string): Promise<void> {
    await this.openMoreActions();
    await this.page.getByRole("menuitem", { name }).click();
  }
}
