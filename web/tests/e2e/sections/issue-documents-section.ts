import type { Locator, Page } from "@playwright/test";

import { Form } from "../components";
import { clickUntilVisible } from "../helpers";
import { DialogMixin, SectionContainerMixin } from "../mixins";

/**
 * Related-documents section on a work item.
 */
export class IssueDocumentsSection extends DialogMixin(
  SectionContainerMixin(Form)
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='issue-documents']")
    );
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

  getLinkButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Link",
      exact: true,
    });
  }

  getDocumentLink(name: string): Locator {
    return this.getSectionContainer().getByRole("link", { name });
  }

  getEmptyState(): Locator {
    return this.getSectionContainer().getByText("No linked documents");
  }

  async openCreateDialog(): Promise<void> {
    await clickUntilVisible(
      this.getAddButton(),
      this.page.getByRole("dialog").filter({ hasText: "Create document" })
    );
    await this.waitForDialog("Create document");
  }

  async openLinkDialog(): Promise<void> {
    await clickUntilVisible(
      this.getLinkButton(),
      this.page.getByRole("dialog").filter({ hasText: "Link document" })
    );
    await this.waitForDialog("Link document");
  }

  async fillTitle(title: string): Promise<void> {
    await this.fillField("Title", title);
  }

  async submitCreate(): Promise<void> {
    await this.page.getByRole("button", { name: "Create document" }).click();
  }

  async linkDocument(title: string): Promise<void> {
    await this.openLinkDialog();
    const dialog = this.page.locator("[data-section='document-link-form']");
    await dialog.getByLabel("Document", { exact: true }).click();
    await this.page
      .locator('[data-slot="command-item"]')
      .filter({ hasText: title })
      .first()
      .click();
    await this.confirmDialog("Link document");
  }
}
