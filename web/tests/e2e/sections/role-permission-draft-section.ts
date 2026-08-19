import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { SectionContainerMixin } from "../mixins";

import type { Action } from "@/lib/api/types";

interface DraftActionOptions {
  action: Action;
}

/**
 * Section for managing bundled actions on the role create page.
 * Roles are capability bundles; they are not granted until a GRANTED edge
 * references them at a scope.
 */
export class RolePermissionDraftSection extends SectionContainerMixin(
  BaseComponent
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='role-permission-draft']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  getActionRow(action: Action): Locator {
    return this.getSectionContainer()
      .getByRole("row")
      .filter({ hasText: action });
  }

  async waitForActionRow(action: Action): Promise<Locator> {
    const row = this.getActionRow(action).first();
    await expect(row).toBeVisible();
    return row;
  }

  async addAction({ action }: DraftActionOptions): Promise<void> {
    const form = this.getSectionContainer().locator("form").first();
    const actionField = form
      .getByRole("combobox")
      .or(form.getByRole("textbox"));
    await actionField.first().click();
    await this.page.getByRole("option", { name: action }).click();
    await form.getByRole("button", { name: /add/i }).click();
    await this.waitForActionRow(action);
  }
}
