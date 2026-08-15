import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

import { Form } from "../components";
import { fillLocator, waitForAnimations } from "../helpers";
import { DialogMixin } from "../mixins";

export type TodoPriorityName = "Normal" | "Important" | "Urgent" | "Critical";

export interface TodoFormFields {
  title: string;
  description?: string;
  priority?: TodoPriorityName;
}

/**
 * Add Todo and Edit Todo dialogs.
 */
export class TodoFormSection extends DialogMixin(Form) {
  constructor(page: Page) {
    super(page);
  }

  getAddDialog(options?: { includeHidden?: boolean }): Locator {
    return this.page.locator("[data-section='todo-add-form']").or(
      this.page.getByRole("dialog", {
        name: "Add Todo",
        includeHidden: options?.includeHidden,
      })
    );
  }

  getEditDialog(): Locator {
    return this.page
      .locator("[data-section='todo-edit-form']")
      .or(this.page.getByRole("dialog", { name: "Edit Todo" }));
  }

  async waitForAddDialog(options?: {
    allowHidden?: boolean;
    timeout?: number;
  }): Promise<Locator> {
    if (options?.allowHidden) {
      await this.page.evaluate(async () => {
        await Promise.all(
          document
            .getAnimations()
            .map((animation) => animation.finished.catch(() => undefined))
        );
      });
    }

    const dialog = this.getAddDialog({ includeHidden: options?.allowHidden });
    await expect(dialog).toBeVisible({ timeout: options?.timeout });
    await waitForAnimations(dialog);
    return dialog;
  }

  async waitForEditDialog(options?: { timeout?: number }): Promise<Locator> {
    const dialog = this.getEditDialog();
    await expect(dialog).toBeVisible({ timeout: options?.timeout });
    await waitForAnimations(dialog);
    return dialog;
  }

  getCreateMoreCheckbox(dialog: Locator): Locator {
    return dialog.getByRole("checkbox", { name: "Create more" });
  }

  getDueDatePicker(dialog: Locator): Locator {
    return dialog.getByLabel("Due Date");
  }

  async setCreateMore(checked: boolean): Promise<void> {
    const dialog = this.getAddDialog({ includeHidden: true });
    const checkbox = this.getCreateMoreCheckbox(dialog);
    if (checked) {
      await checkbox.check();
    } else {
      await checkbox.uncheck();
    }
  }

  async fillTodoFields(
    dialog: Locator,
    fields: TodoFormFields,
    options?: { force?: boolean }
  ): Promise<void> {
    await fillLocator(dialog.getByLabel("Title"), fields.title);
    if (fields.description !== undefined) {
      await fillLocator(dialog.getByLabel("Description"), fields.description);
    }
    if (fields.priority) {
      await dialog
        .getByRole("combobox")
        .click(options?.force ? { force: true } : undefined);
      await this.page.getByRole("option", { name: fields.priority }).click();
    }
  }

  async selectDueDate(dialog: Locator, date: Date): Promise<void> {
    await this.getDueDatePicker(dialog).click();
    const calendar = this.page.locator("[data-slot='calendar']");
    await expect(calendar).toBeVisible();
    await waitForAnimations(calendar);

    const dataDay = await this.page.evaluate(
      (iso) => new Date(iso).toLocaleDateString(),
      date.toISOString()
    );
    const dayButton = calendar.locator(`[data-day="${dataDay}"]`);

    for (let attempt = 0; attempt < 12; attempt++) {
      if (await dayButton.isVisible()) {
        await dayButton.click();
        return;
      }
      await this.page
        .getByRole("button", { name: "Go to the Next Month" })
        .click();
    }

    throw new Error(`Could not find due date ${dataDay} in the calendar`);
  }

  async submitAdd(options?: {
    force?: boolean;
    includeHidden?: boolean;
  }): Promise<void> {
    const dialog = this.getAddDialog({ includeHidden: options?.includeHidden });
    await dialog
      .getByRole("button", {
        name: "Add todo",
        includeHidden: options?.includeHidden || undefined,
      })
      .click(options?.force ? { force: true } : undefined);
  }

  async submitEdit(): Promise<void> {
    await this.getEditDialog()
      .getByRole("button", { name: "Update todo" })
      .click();
  }

  async fillAndSubmitAdd(
    fields: TodoFormFields,
    options?: { allowHidden?: boolean }
  ): Promise<void> {
    const allowHidden = options?.allowHidden ?? false;
    const dialog = await this.waitForAddDialog({ allowHidden });
    await this.fillTodoFields(dialog, fields, { force: allowHidden });
    await this.submitAdd({ force: allowHidden, includeHidden: allowHidden });
  }
}
