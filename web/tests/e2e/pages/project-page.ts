import type { Locator, Page } from "@playwright/test";

import { Form } from "../components";
import { BaseComponent } from "../components/base";
import { clickUntilVisible, navigateAndWait } from "../helpers";
import { DialogMixin, SectionContainerMixin } from "../mixins";

class ProjectDocumentsSection extends DialogMixin(SectionContainerMixin(Form)) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(this.page.locator("[data-section='documents']"));
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
    return this.getSectionContainer().getByText("No related documents");
  }

  async openCreateDialog(): Promise<void> {
    await clickUntilVisible(
      this.getAddButton(),
      this.page.getByRole("dialog").filter({ hasText: "Create document" })
    );
    await this.waitForDialog("Create document");
  }

  async fillTitle(title: string): Promise<void> {
    await this.fillField("Title", title);
  }

  async submitCreate(): Promise<void> {
    await this.page.getByRole("button", { name: "Create document" }).click();
  }

  async linkDocument(title: string): Promise<void> {
    await clickUntilVisible(
      this.getLinkButton(),
      this.page.getByRole("dialog").filter({ hasText: "Link document" })
    );
    await this.waitForDialog("Link document");
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

/**
 * Project overview at /namespaces/:namespaceId/projects/:projectId.
 */
export class ProjectPage extends BaseComponent {
  public readonly documents: ProjectDocumentsSection;

  constructor(page: Page) {
    super(page);
    this.documents = new ProjectDocumentsSection(page);
  }

  async goto(namespaceId: string, projectId: string): Promise<void> {
    await navigateAndWait(
      this.page,
      `/namespaces/${namespaceId}/projects/${projectId}`
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.documents.waitForLoad(options);
  }
}
