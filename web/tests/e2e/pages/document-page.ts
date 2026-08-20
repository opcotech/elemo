import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { clickUntilVisible, navigateAndWait } from "../helpers";
import { DocumentEditorSection } from "../sections/document-editor-section";

/**
 * Page Object Model for a document at /documents/:id.
 */
export class DocumentPage extends BaseComponent {
  public readonly editor: DocumentEditorSection;

  constructor(page: Page) {
    super(page);
    this.editor = new DocumentEditorSection(page);
  }

  async goto(documentId: string): Promise<void> {
    await navigateAndWait(this.page, `/documents/${documentId}`);
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.editor.waitForLoad(options);
  }

  getLocation(): Locator {
    return this.page.locator("[data-section='document-location']");
  }

  async openMoveDialog(): Promise<void> {
    await this.editor.clickMoreAction("Move");
  }

  async openChangeLibraryDialog(): Promise<void> {
    await this.editor.clickMoreAction("Change library");
    await this.page.getByRole("dialog", { name: "Change library" }).waitFor();
  }

  async changeLibrary(libraryLabel: string): Promise<void> {
    await this.openChangeLibraryDialog();
    const dialog = this.page.getByRole("dialog", { name: "Change library" });
    const trigger = dialog.getByLabel("Library", { exact: true });
    const listbox = this.page.getByRole("listbox");
    const option = this.page.getByRole("option", {
      name: libraryLabel,
      exact: true,
    });
    await clickUntilVisible(trigger, listbox);
    await option.click();
    try {
      await expect(trigger).toContainText(libraryLabel, { timeout: 2_000 });
    } catch {
      await clickUntilVisible(trigger, listbox);
      await option.click();
      await expect(trigger).toContainText(libraryLabel);
    }
    await dialog.getByRole("button", { name: "Move to library" }).click();
  }

  async openDeleteDialog(): Promise<void> {
    await this.editor.clickMoreAction("Delete");
  }

  async confirmDelete(): Promise<void> {
    await this.page
      .getByRole("alertdialog")
      .getByRole("button", { name: "Delete" })
      .click();
  }

  async cancelDelete(): Promise<void> {
    await this.page.getByRole("button", { name: "Cancel" }).click();
  }
}
