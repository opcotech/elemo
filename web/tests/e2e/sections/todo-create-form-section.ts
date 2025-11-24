import type { Page } from "@playwright/test";

import { Form } from "../components";
import { BaseComponent } from "../components/base";
import { waitForElementVisible } from "../helpers";

export class TodoCreateFormSection extends BaseComponent {
  private form: Form;

  constructor(page: Page) {
    super(page);
    this.form = new Form(page);
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await waitForElementVisible(
      this.page.getByRole("heading", { name: "Add Todo" }),
      options
    );
  }

  async fillTodoFields(fields: {
    title: string;
    description?: string;
    priority?: "normal" | "important" | "urgent" | "critical";
    dueDate?: Date;
  }): Promise<void> {
    await this.form.fillFields({ Title: fields.title });
    
    if (fields.description) {
      await this.page.getByPlaceholder(/Enter todo description/i).fill(fields.description);
    }
    
    if (fields.priority) {
      await this.page.getByRole("combobox", { name: /priority/i }).click();
      await this.page.getByRole("option", { name: fields.priority, exact: true }).click();
    }
    
    if (fields.dueDate) {
      const dateButton = this.page.getByRole("button", { name: /due date/i });
      await dateButton.click();
      // Select date from calendar picker
      const formattedDate = fields.dueDate.getDate().toString();
      await this.page.getByRole("gridcell", { name: formattedDate, exact: true }).click();
    }
  }

  async setCreateMore(checked: boolean): Promise<void> {
    const checkbox = this.page.getByRole("checkbox", { name: /create more/i });
    const isChecked = await checkbox.isChecked();
    if (checked !== isChecked) {
      await checkbox.click();
    }
  }

  async submit(): Promise<void> {
    await this.form.submit("Add todo");
  }

  async createTodo(fields: {
    title: string;
    description?: string;
    priority?: "normal" | "important" | "urgent" | "critical";
    dueDate?: Date;
    createMore?: boolean;
  }): Promise<void> {
    await this.fillTodoFields(fields);
    if (fields.createMore) {
      await this.setCreateMore(true);
    }
    await this.submit();
  }
}