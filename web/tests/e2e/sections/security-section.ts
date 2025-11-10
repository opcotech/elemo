import type { Page } from "@playwright/test";

import { Form } from "../components";
import { BaseComponent } from "../components/base";
import { waitForElementVisible } from "../helpers";

/**
 * Reusable Security Section component.
 * Handles password change form interactions.
 * Can be composed into any page that displays security settings.
 */
export class SecuritySection extends BaseComponent {
  private form: Form;

  constructor(page: Page) {
    super(page);
    this.form = new Form(page);
  }

  /**
   * Wait for section to load and be visible.
   */
  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await waitForElementVisible(
      this.page.getByRole("heading", {
        name: "Password & Authentication",
      }),
      options
    );
  }

  /**
   * Get the form component for advanced operations.
   */
  getForm(): Form {
    return this.form;
  }

  /**
   * Fill password change form fields.
   */
  async fillPasswordFields(fields: {
    currentPassword: string;
    newPassword: string;
    confirmPassword: string;
  }): Promise<void> {
    await this.form.fillFields({
      "Current Password": fields.currentPassword,
      "New Password": fields.newPassword,
      "Confirm New Password": fields.confirmPassword,
    });
  }

  /**
   * Submit the password change form.
   */
  async submitPasswordChange(): Promise<void> {
    await this.form.submit("Update Password");
  }

  /**
   * Get the current password field locator.
   */
  getCurrentPasswordField() {
    return this.page.getByLabel("Current Password", { exact: true });
  }

  /**
   * Get the new password field locator.
   */
  getNewPasswordField() {
    return this.page.getByLabel("New Password", { exact: true });
  }

  /**
   * Get the confirm password field locator.
   */
  getConfirmPasswordField() {
    return this.page.getByLabel("Confirm New Password", { exact: true });
  }

  /**
   * Get the submit button locator.
   */
  getSubmitButton() {
    return this.page.getByRole("button", { name: "Update Password" });
  }

  /**
   * Get all password visibility toggle buttons.
   */
  getPasswordToggleButtons() {
    return this.page.getByRole("button", {
      name: /Show password|Hide password/,
    });
  }

  /**
   * Toggle password visibility for a specific field.
   */
  async togglePasswordVisibility(fieldIndex: number): Promise<void> {
    const toggles = this.getPasswordToggleButtons();
    await toggles.nth(fieldIndex).click();
  }
}
