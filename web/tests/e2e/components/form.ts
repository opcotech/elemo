import {
  clearFormField,
  fillFormField,
  submitForm,
  waitForFormSubmission,
} from "../helpers";
import { BaseComponent } from "./base";

/**
 * Reusable Form component helper.
 * Provides common form operations with a composable API.
 */
export class Form extends BaseComponent {
  /**
   * Clear a form field by label.
   */
  async clearField(label: string): Promise<void> {
    await clearFormField(this.page, label);
  }

  /**
   * Fill a form field by label.
   */
  async fillField(label: string, value: string): Promise<void> {
    await fillFormField(this.page, label, value);
  }

  /**
   * Fill multiple fields at once.
   */
  async fillFields(fields: Record<string, string>): Promise<void> {
    for (const [label, value] of Object.entries(fields)) {
      await this.fillField(label, value);
    }
  }

  /**
   * Submit the form by clicking a button.
   */
  async submit(buttonText: string): Promise<void> {
    await submitForm(this.page, buttonText);
    await waitForFormSubmission(this.page);
  }

  /**
   * Get a field locator by label.
   * Useful for composing with other components.
   */
  getField(label: string) {
    return this.page.getByLabel(label, { exact: true });
  }

  /**
   * Get the submit button locator.
   */
  getSubmitButton(buttonText: string) {
    return this.page.getByRole("button", { name: buttonText });
  }
}
