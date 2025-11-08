import {
  fillFormField,
  submitForm,
  waitForFormSubmission,
} from "../helpers/forms";
import type { Page } from "@playwright/test";

/**
 * Reusable Form component helper.
 * Provides common form operations.
 */
export class Form {
  constructor(private page: Page) {}

  /**
   * Fill a form field by label.
   */
  async fillField(label: string, value: string): Promise<void> {
    await fillFormField(this.page, label, value);
  }

  /**
   * Submit the form.
   */
  async submit(buttonText: string): Promise<void> {
    await submitForm(this.page, buttonText);
    await waitForFormSubmission(this.page);
  }

  /**
   * Fill multiple fields at once.
   */
  async fillFields(fields: Record<string, string>): Promise<void> {
    for (const [label, value] of Object.entries(fields)) {
      await this.fillField(label, value);
    }
  }
}
