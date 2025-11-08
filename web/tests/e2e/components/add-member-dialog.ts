import { expect } from "@playwright/test";
import { waitForDropdownOpen } from "../helpers/elements";
import { Dialog } from "./dialog";
import type { Page } from "@playwright/test";

/**
 * Reusable Add Member to Role Dialog component helper.
 * Provides methods to interact with the add member to role dialog.
 */
export class AddMemberDialog extends Dialog {
  constructor(page: Page) {
    super(page);
  }

  /**
   * Wait for the add member dialog to appear.
   */
  async waitFor(): Promise<void> {
    await super.waitFor("Add Member to Role");
  }

  /**
   * Select a member from the member dropdown.
   * @param memberName - Name of the member to select (e.g., "John Doe")
   */
  async selectMember(memberName: string): Promise<void> {
    const memberCombobox = this.page.getByLabel("Select Member", {
      exact: true,
    });
    await memberCombobox.click();
    await waitForDropdownOpen(memberCombobox);
    const memberOption = this.page.getByRole("option", { name: memberName });
    await memberOption.click();
  }

  /**
   * Add the selected member to the role.
   */
  async addMember(): Promise<void> {
    const addButton = this.page.getByRole("button", { name: "Add Member" });
    await addButton.click();
  }

  /**
   * Select a member and add them to the role.
   * @param memberName - Name of the member to add
   */
  async addMemberToRole(memberName: string): Promise<void> {
    await this.selectMember(memberName);
    await this.addMember();
  }
}
