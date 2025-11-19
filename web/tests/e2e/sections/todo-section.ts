import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { waitForElementVisible } from "../helpers";

export class TodoSection extends BaseComponent {
  constructor(page: Page) {
    super(page);
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await waitForElementVisible(
      this.page.getByRole("heading", { name: "Todo Items" }),
      options
    );
  }

  getTodoSheet(): Locator {
    return this.page.getByRole("dialog").filter({ hasText: "Todo Items" });
  }

  async isOpen(): Promise<boolean> {
    return await this.getTodoSheet().isVisible();
  }

  async close(): Promise<void> {
    await this.page.keyboard.press("Escape");
  }

  getAddTodoButton(): Locator {
    return this.page.getByRole("button", { name: "Add Todo" });
  }

  async clickAddTodo(): Promise<void> {
    await this.getAddTodoButton().click();
  }

  getTodoItems(): Locator {
    return this.getTodoSheet().locator('[class*="group"]').filter({ hasText: /^(?!No todos found)/ });
  }

  getTodoByTitle(title: string): Locator {
    return this.getTodoSheet().locator('[class*="group"]').filter({ hasText: title }).first();
  }

  async getTodoCount(): Promise<number> {
    return await this.getTodoItems().count();
  }

  getEmptyState(): Locator {
    return this.getTodoSheet().getByText("No todos found");
  }
}
