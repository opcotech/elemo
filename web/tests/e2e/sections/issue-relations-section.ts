import type { Locator, Page } from "@playwright/test";

import { Form } from "../components";
import { clickUntilVisible } from "../helpers";
import { DialogMixin, SectionContainerMixin } from "../mixins";

/**
 * Issue relations section: add, change kind, and remove relations.
 */
export class IssueRelationsSection extends DialogMixin(
  SectionContainerMixin(Form)
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='issue-relations']")
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

  async clickAdd(): Promise<void> {
    await clickUntilVisible(
      this.getAddButton(),
      this.page.getByRole("dialog").filter({ hasText: "Add relation" })
    );
    await this.waitForDialog("Add relation");
  }

  getRelation(relatedKey: string): Locator {
    return this.getSectionContainer()
      .getByRole("listitem")
      .filter({ hasText: relatedKey });
  }

  getRelationKindButton(relatedKey: string): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: `Relation kind for ${relatedKey}`,
    });
  }

  getRemoveRelationButton(relatedKey: string): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: `Remove relation to ${relatedKey}`,
    });
  }

  async addRelation(fields: { kind: string; issue: string }): Promise<void> {
    await this.clickAdd();
    const dialog = this.page.locator(
      "[data-section='issue-relation-add-form']"
    );
    await dialog.getByLabel("Kind", { exact: true }).click();
    await this.page.getByRole("option", { name: fields.kind }).click();
    await dialog.getByLabel("Issue", { exact: true }).click();
    await this.page
      .locator('[data-slot="command-item"]')
      .filter({ hasText: fields.issue })
      .first()
      .click();
    await this.confirmDialog("Add relation");
  }

  async changeKind(relatedKey: string, kind: string): Promise<void> {
    await this.getRelationKindButton(relatedKey).click();
    await this.page.getByRole("menuitemradio", { name: kind }).click();
  }

  async removeRelation(relatedKey: string): Promise<void> {
    await this.getRemoveRelationButton(relatedKey).click();
  }
}
