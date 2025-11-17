import type { Page } from "@playwright/test";

import { Form } from "../components";
import { BaseComponent } from "../components/base";
import { waitForElementVisible } from "../helpers";

export class TodoEditFormSection extends BaseComponent {
  private form: Form;

  constructor(page: Page) {
    super(page);
    this.form = new Form(page);
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await waitForElementVisible(
      this.page.getByRole("heading", { name: "Edit Todo" }),
      options
    );
  }

  async fillTodoFields(fields: {
    title?: string;
    description?: string;
    priority?: "normal" | "important" | "urgent" | "critical";
    dueDate?: Date | null;
  }): Promise<void> {
    if (fields.title !== undefined) {
      const titleInput = this.page.getByLabel(/title/i);
      await titleInput.clear();
      await titleInput.fill(fields.title);
    }
    
    if (fields.description !== undefined) {
      const descInput = this.page.getByPlaceholder(/Enter todo description/i);
      await descInput.clear();
      await descInput.fill(fields.description);
    }
    
    if (fields.priority) {
      await this.page.getByRole("combobox", { name: /priority/i }).click();
      await this.page.getByRole("option", { name: fields.priority, exact: true }).click();
    }
    
    if (fields.dueDate !== undefined) {
      if (fields.dueDate === null) {
        // Clear date
        const dateButton = this.page.getByRole("button", { name: /due date/i });
        const clearButton = dateButton.locator('[aria-label="Clear"]');
        if (await clearButton.isVisible()) {
          await clearButton.click();
        }
      } else {
        const dateButton = this.page.getByRole("button", { name: /due date/i });
        await dateButton.click();
        const formattedDate = fields.dueDate.getDate().toString();
        await this.page.getByRole("gridcell", { name: formattedDate, exact: true }).click();
      }
    }
  }

  async submit(): Promise<void> {
    await this.form.submit("Update todo");
  }

  async updateTodo(fields: {
    title?: string;
    description?: string;
    priority?: "normal" | "important" | "urgent" | "critical";
    dueDate?: Date | null;
  }): Promise<void> {
    await this.fillTodoFields(fields);
    await this.submit();
  }
}