import { expect } from "@playwright/test";
import type { Page } from "@playwright/test";

import { Form } from "../components";
import { BaseComponent } from "../components/base";
import { waitForElementVisible } from "../helpers";

/**
 * Reusable Login Section component.
 * Handles login form interactions.
 * Can be composed into any page that displays login functionality.
 */
export class LoginSection extends BaseComponent {
  private form: Form;

  constructor(page: Page) {
    super(page);
    this.form = new Form(page);
  }

  /**
   * Wait for section to load and be visible.
   */
  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    const signIn = this.page.getByRole("button", { name: "Sign in" });
    await waitForElementVisible(signIn, options);
    // Client mount enables the button; avoid native GET submit pre-hydration.
    await expect(signIn).toBeEnabled({ timeout: options?.timeout ?? 15_000 });
  }

  /**
   * Get the form component for advanced operations.
   */
  getForm(): Form {
    return this.form;
  }

  /**
   * Fill login form fields.
   */
  async fillLoginFields(fields: {
    email: string;
    password: string;
  }): Promise<void> {
    await this.form.fillFields({
      Email: fields.email,
      Password: fields.password,
    });
  }

  /**
   * Submit the login form.
   */
  async submit(): Promise<void> {
    await this.form.submit("Sign in");
  }

  /**
   * Perform complete login flow.
   */
  async login(email: string, password: string): Promise<void> {
    await this.fillLoginFields({ email, password });
    await this.submit();
  }
}
