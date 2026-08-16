import type { Locator, Page } from "@playwright/test";

import { Form } from "../components";
import { DialogMixin, SectionContainerMixin } from "../mixins";

export type QuickCreateEntityType = "Personal todo" | "Work item" | "Document";

/**
 * Quick create dialog: entity type plus work and todo fields.
 */
export class QuickCreateSection extends DialogMixin(
  SectionContainerMixin(Form)
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='quick-create']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForDialog("Quick create", options);
    await this.waitForContainerLoad(options);
  }

  getEntityTypeSelect(): Locator {
    return this.getSectionContainer().getByLabel("Entity type");
  }

  async selectEntityType(type: QuickCreateEntityType): Promise<void> {
    await this.getEntityTypeSelect().click();
    await this.page.getByRole("option", { name: type, exact: true }).click();
  }

  getMorePropertiesTrigger(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "More properties",
    });
  }

  async expandMoreProperties(): Promise<void> {
    const trigger = this.getMorePropertiesTrigger();
    if ((await trigger.getAttribute("aria-expanded")) !== "true") {
      await trigger.click();
    }
  }

  getCreateIssueButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Create issue",
    });
  }

  getCreateTodoButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Create todo",
    });
  }

  getCreateUnavailableButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Create unavailable",
    });
  }

  async fillWorkTitle(title: string): Promise<void> {
    await this.fillField("Title", title);
  }

  async fillWorkKind(kind: string): Promise<void> {
    await this.expandMoreProperties();
    await this.getSectionContainer().getByLabel("Kind").click();
    await this.page.getByRole("option", { name: kind, exact: true }).click();
  }

  async fillWorkDescription(description: string): Promise<void> {
    await this.expandMoreProperties();
    await this.fillField("Description", description);
  }

  async submitWork(): Promise<void> {
    await this.getCreateIssueButton().click();
  }

  async submitTodo(): Promise<void> {
    await this.getCreateTodoButton().click();
  }

  async cancel(): Promise<void> {
    await this.cancelDialog();
  }
}
