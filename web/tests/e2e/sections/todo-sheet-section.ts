import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { waitForAnimations } from "../helpers";
import { DialogMixin, SectionContainerMixin } from "../mixins";

/**
 * Todo sheet: groups, items, and open/close actions extracted from todos.spec.ts.
 */
export class TodoSheetSection extends DialogMixin(
  SectionContainerMixin(BaseComponent)
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(this.page.locator("[data-section='todo-sheet']"));
  }

  getSheet(): Locator {
    return this.page.getByRole("dialog", { name: "Todo Items" });
  }

  getItem(title: string): Locator {
    return this.getSheet().getByRole("listitem").filter({ hasText: title });
  }

  getGroup(label: string): Locator {
    return this.getSheet().getByRole("list", { name: `${label} todos` });
  }

  async waitForOpen(options?: { timeout?: number }): Promise<void> {
    await expect(this.getSheet()).toBeVisible({ timeout: options?.timeout });
    await waitForAnimations(this.getSheet());
  }

  async openFromHeader(): Promise<Locator> {
    await this.page.getByRole("button", { name: "Show todo list" }).click();
    await this.waitForOpen();
    return this.getSheet();
  }

  async closeWithButton(): Promise<void> {
    await this.getSheet().getByRole("button", { name: "Close" }).click();
    await expect(this.getSheet()).not.toBeVisible();
  }

  async closeWithEscape(): Promise<void> {
    await this.page.keyboard.press("Escape");
    await expect(this.getSheet()).not.toBeVisible();
  }

  getAddButton(): Locator {
    return this.getSheet().getByRole("button", { name: "Add Todo" }).first();
  }

  getEmptyStateAddButton(): Locator {
    return this.getSheet()
      .locator("[data-slot='empty']")
      .getByRole("button", { name: "Add Todo" });
  }

  async clickAddTodo(): Promise<void> {
    await this.getAddButton().click();
  }

  async clickEmptyStateAddTodo(): Promise<void> {
    await this.getEmptyStateAddButton().click();
  }

  async clickEditItem(title: string): Promise<void> {
    const item = this.getItem(title);
    await item.hover();
    await item.getByRole("button", { name: "Edit todo" }).click();
  }

  async clickDeleteItem(title: string): Promise<void> {
    const item = this.getItem(title);
    await item.hover();
    await item.getByRole("button", { name: "Delete todo" }).click();
  }

  getCompleteCheckbox(title: string): Locator {
    return this.getItem(title).getByRole("checkbox", {
      name: `Mark "${title}" as complete`,
    });
  }

  getIncompleteCheckbox(title: string): Locator {
    return this.getItem(title).getByRole("checkbox", {
      name: `Mark "${title}" as incomplete`,
    });
  }

  async markComplete(title: string): Promise<void> {
    const item = this.getItem(title);
    await item.hover();
    const checkbox = this.getCompleteCheckbox(title);
    await expect(checkbox).toBeEnabled();
    await checkbox.click();
  }

  async markIncomplete(title: string): Promise<void> {
    const item = this.getItem(title);
    await item.hover();
    const checkbox = this.getIncompleteCheckbox(title);
    await expect(checkbox).toBeEnabled();
    await checkbox.click();
  }
}
