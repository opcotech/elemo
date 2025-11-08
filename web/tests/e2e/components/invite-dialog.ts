import { expect } from "@playwright/test";
import { waitForDropdownOpen } from "../helpers/elements";
import { Dialog } from "./dialog";
import type { Page } from "@playwright/test";

/**
 * Reusable Invite Member Dialog component helper.
 * Provides methods to interact with the invite member dialog.
 */
export class InviteDialog extends Dialog {
  constructor(page: Page) {
    super(page);
  }

  /**
   * Wait for the invite dialog to appear.
   */
  async waitFor(): Promise<void> {
    await super.waitFor("Invite Member");
  }

  /**
   * Fill the email address field.
   */
  async fillEmail(email: string): Promise<void> {
    const emailField = this.page.getByLabel("Email Address");
    await emailField.fill(email);
  }

  /**
   * Select a role from the role dropdown.
   * @param roleName - Name of the role to select
   */
  async selectRole(roleName: string): Promise<void> {
    const roleCombobox = this.page.getByLabel("Role (Optional)");
    await roleCombobox.click();
    await waitForDropdownOpen(roleCombobox);
    const roleOption = this.page.getByRole("option", { name: roleName });
    await roleOption.click();
  }

  /**
   * Send the invitation.
   */
  async sendInvitation(): Promise<void> {
    const sendButton = this.page.getByRole("button", {
      name: "Send Invitation",
    });
    await sendButton.click();
  }

  /**
   * Fill the invite form and send.
   * @param email - Email address to invite
   * @param roleName - Optional role name to assign
   */
  async inviteMember(email: string, roleName?: string): Promise<void> {
    await this.fillEmail(email);
    if (roleName) {
      await this.selectRole(roleName);
    }
    await this.sendInvitation();
  }
}
