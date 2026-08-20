import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { clickUntilVisible } from "../helpers";
import { EmptyStateMixin, SearchMixin, SectionContainerMixin } from "../mixins";

/**
 * Organization, namespace, and project document lists.
 */
export class DocumentListSection extends SectionContainerMixin(
  SearchMixin(EmptyStateMixin(BaseComponent))
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(this.page.locator("[data-section='documents']"));
    this.setSearchConfig({
      getSectionContainer: () => this.getSectionContainer(),
      searchPlaceholder: "Search documents...",
    });
    this.setEmptyStateConfig({
      emptyStateText: "No documents",
      getSectionContainer: () => this.getSectionContainer(),
    });
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
    const container = this.getSectionContainer();
    const loading = container.locator('[role="status"][aria-busy="true"]');
    const ready = container
      .locator('[role="listitem"]')
      .or(container.getByText("No documents yet"))
      .or(container.getByText("No documents in this folder"))
      .or(container.getByText("No related documents"))
      .or(container.getByText("No document matches"))
      .or(container.getByText("Limited access"));
    await expect(async () => {
      await expect(loading).toHaveCount(0);
      await expect(ready.first()).toBeVisible();
    }).toPass({ timeout: options?.timeout ?? 15_000 });
  }

  getCreateButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Create",
      exact: true,
    });
  }

  getNewFolderButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "New folder",
    });
  }

  getBrowseMenuButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Browse documents",
    });
  }

  getLibraryLink(): Locator {
    return this.page.getByRole("menuitemradio", {
      name: "Library",
      exact: true,
    });
  }

  getAllLink(): Locator {
    return this.page.getByRole("menuitemradio", {
      name: "All",
      exact: true,
    });
  }

  getFolderLink(name: string): Locator {
    return this.getSectionContainer()
      .locator("[data-section='library-folders']")
      .getByRole("link", { name });
  }

  getDocumentActionsButton(title: string): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: `Document actions for ${title}`,
    });
  }

  getFolderActionsButton(name: string): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: `Folder actions for ${name}`,
    });
  }

  getSortButton(): Locator {
    return this.getSectionContainer().getByLabel("Sort documents");
  }

  getCreatorButton(): Locator {
    return this.getSectionContainer().getByLabel("Filter by creator");
  }

  getDocumentLink(name: string): Locator {
    return this.getSectionContainer()
      .locator("a")
      .filter({ has: this.page.locator("h2", { hasText: name }) });
  }

  getDocumentTitles(): Locator {
    return this.getSectionContainer().locator("h2");
  }

  getMenu(): Locator {
    return this.page
      .locator('[data-slot="dropdown-menu-content"]')
      .filter({ visible: true });
  }

  async clickCreate(): Promise<void> {
    await this.getCreateButton().first().click();
  }

  async clickDocument(name: string): Promise<void> {
    await this.getDocumentLink(name).first().click();
  }

  async clickFolder(name: string): Promise<void> {
    await this.getFolderLink(name).click();
  }

  async clickAll(): Promise<void> {
    await clickUntilVisible(this.getBrowseMenuButton(), this.getAllLink());
    await this.getAllLink().click();
  }

  async clickLibrary(): Promise<void> {
    await clickUntilVisible(this.getBrowseMenuButton(), this.getLibraryLink());
    await this.getLibraryLink().click();
  }

  async openDocumentActions(title: string): Promise<void> {
    await clickUntilVisible(
      this.getDocumentActionsButton(title),
      this.getMenu()
    );
  }

  async openMove(documentTitle: string): Promise<void> {
    await this.openDocumentActions(documentTitle);
    await this.page.getByRole("menuitem", { name: "Move" }).click();
  }

  async openFolderActions(name: string): Promise<void> {
    await clickUntilVisible(this.getFolderActionsButton(name), this.getMenu());
  }

  async selectSort(
    label: "Updated" | "Created" | "Oldest" | "Title"
  ): Promise<void> {
    const option = this.page.getByRole("menuitemradio", { name: label });
    await clickUntilVisible(this.getSortButton(), option);
    await option.click();
  }

  async selectCreator(label: string): Promise<void> {
    const search = this.page.getByPlaceholder("Search people…");
    await clickUntilVisible(this.getCreatorButton(), search);
    await search.click();
    await search.press("ControlOrMeta+A");
    await search.press("Backspace");
    if (label !== "Anyone") {
      await search.pressSequentially(label, { delay: 10 });
    }
    await this.page
      .getByRole("listbox")
      .getByRole("option", { name: label })
      .click();
  }
}
