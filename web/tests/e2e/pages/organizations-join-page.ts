import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait, waitForSuccessToast } from "../helpers";

/**
 * Page object for the public /organizations/join invite acceptance flow.
 */
export class OrganizationsJoinPage extends BaseComponent {
  constructor(page: Page) {
    super(page);
  }

  async goto(organizationId: string, token: string): Promise<void> {
    const searchParams = new URLSearchParams({
      organization: organizationId,
      token,
    });
    await navigateAndWait(this.page, `/organizations/join?${searchParams}`);
  }

  private getAcceptButton() {
    return this.page.getByRole("button", { name: /accept invitation/i });
  }

  async acceptInvitation(options?: { password?: string }): Promise<void> {
    const passwordField = this.page.getByLabel("Password");
    const confirmPasswordField = this.page.getByLabel("Confirm Password");
    const fillPasswordsAndSubmit = async () => {
      const password = options?.password;
      if (!password) {
        throw new Error("Password is required to accept this invitation");
      }
      await passwordField.fill(password);
      await confirmPasswordField.fill(password);
      await this.getAcceptButton().click();
    };

    if (await passwordField.isVisible().catch(() => false)) {
      await fillPasswordsAndSubmit();
    } else {
      await this.getAcceptButton().click();
      const passwordAppeared = await passwordField
        .waitFor({ state: "visible", timeout: 2000 })
        .then(() => true)
        .catch(() => false);
      if (passwordAppeared) {
        await fillPasswordsAndSubmit();
      }
    }

    await waitForSuccessToast(this.page, "Invitation accepted");
  }
}
