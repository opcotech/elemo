import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { clickUntilVisible, fillLocator, navigateAndWait } from "../helpers";
import { DocumentListSection } from "../sections/document-list-section";
import { QuickCreateSection } from "../sections/quick-create-section";

/**
 * Page Object Model for org, namespace, and project document lists.
 */
export class DocumentsListPage extends BaseComponent {
  public readonly list: DocumentListSection;
  public readonly quickCreate: QuickCreateSection;

  constructor(page: Page) {
    super(page);
    this.list = new DocumentListSection(page);
    this.quickCreate = new QuickCreateSection(page);
  }

  async gotoHub(): Promise<void> {
    await navigateAndWait(this.page, "/documents", {
      ready: this.page.locator("[data-section='document-hub']"),
    });
  }

  getHubLibraryLink(name: string): Locator {
    return this.page
      .locator("[data-section='document-libraries']")
      .getByRole("link", { name });
  }

  async gotoOrganization(organizationSlug: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `/organizations/${organizationSlug}/documents`
    );
  }

  async gotoNamespace(
    organizationSlug: string,
    namespaceSlug: string,
    search?: { folder?: string; all?: boolean }
  ): Promise<void> {
    const params = new URLSearchParams();
    if (search?.all) {
      params.set("all", "true");
    } else if (search?.folder) {
      params.set("folder", search.folder);
    }
    const query = params.toString();
    await navigateAndWait(
      this.page,
      `/organizations/${organizationSlug}/namespaces/${namespaceSlug}/documents${query ? `?${query}` : ""}`
    );
  }

  async gotoProject(
    organizationSlug: string,
    namespaceSlug: string,
    projectKey: string
  ): Promise<void> {
    await navigateAndWait(
      this.page,
      `/organizations/${organizationSlug}/namespaces/${namespaceSlug}/projects/${projectKey}/documents`
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.list.waitForLoad(options);
  }

  async createFolder(name: string): Promise<void> {
    await this.list.getNewFolderButton().click();
    const dialog = this.page.getByRole("dialog", { name: "New folder" });
    await fillLocator(dialog.getByLabel("Name"), name);
    await dialog.getByRole("button", { name: "Create folder" }).click();
  }

  async selectFolder(folderLabel: string): Promise<void> {
    const dialog = this.page.getByRole("dialog").filter({ visible: true });
    const trigger = dialog.getByLabel("Folder", { exact: true });
    const search = this.page.getByPlaceholder("Search folders…");
    await clickUntilVisible(trigger, search);
    await fillLocator(search, folderLabel);
    await this.page
      .getByRole("option", { name: folderLabel, exact: true })
      .click();
  }

  async moveDocument(
    documentTitle: string,
    folderLabel: string
  ): Promise<void> {
    await this.list.openMove(documentTitle);
    const dialog = this.page.getByRole("dialog", { name: "Move document" });
    await this.selectFolder(folderLabel);
    await dialog.getByRole("button", { name: "Move", exact: true }).click();
  }

  async renameDocument(title: string, nextTitle: string): Promise<void> {
    await this.list.openDocumentActions(title);
    await this.page.getByRole("menuitem", { name: "Rename" }).click();
    const dialog = this.page.getByRole("dialog", { name: "Rename document" });
    await fillLocator(dialog.getByLabel("Title"), nextTitle);
    await dialog.getByRole("button", { name: "Rename" }).click();
  }

  async deleteDocument(title: string): Promise<void> {
    await this.list.openDocumentActions(title);
    await this.page.getByRole("menuitem", { name: "Delete" }).click();
    const dialog = this.page.getByRole("alertdialog", {
      name: `Are you sure you want to delete ${title}?`,
    });
    await dialog.getByRole("button", { name: "Delete" }).click();
  }

  async renameFolder(name: string, nextName: string): Promise<void> {
    await this.list.openFolderActions(name);
    await this.page.getByRole("menuitem", { name: "Rename" }).click();
    const dialog = this.page.getByRole("dialog", { name: "Rename folder" });
    await fillLocator(dialog.getByLabel("Name"), nextName);
    await dialog.getByRole("button", { name: "Rename" }).click();
  }

  async moveFolder(name: string, folderLabel: string): Promise<void> {
    await this.list.openFolderActions(name);
    await this.page.getByRole("menuitem", { name: "Move" }).click();
    const dialog = this.page.getByRole("dialog", { name: "Move folder" });
    await this.selectFolder(folderLabel);
    await dialog.getByRole("button", { name: "Move", exact: true }).click();
  }

  async deleteFolder(name: string): Promise<void> {
    await this.list.openFolderActions(name);
    await this.page.getByRole("menuitem", { name: "Delete" }).click();
    const dialog = this.page.getByRole("alertdialog", {
      name: `Are you sure you want to delete ${name}?`,
    });
    await dialog.getByRole("button", { name: "Delete" }).click();
  }

  async openFolderMovePicker(name: string): Promise<void> {
    await this.list.openFolderActions(name);
    await this.page.getByRole("menuitem", { name: "Move" }).click();
    await this.page.getByRole("dialog", { name: "Move folder" }).waitFor();
    await clickUntilVisible(
      this.page
        .getByRole("dialog", { name: "Move folder" })
        .getByLabel("Folder", { exact: true }),
      this.page.getByPlaceholder("Search folders…")
    );
  }
}
