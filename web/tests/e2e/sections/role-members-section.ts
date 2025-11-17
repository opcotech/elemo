import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { waitForSuccessToast } from "../helpers";
import {
  DialogMixin,
  EmptyStateMixin,
  SectionContainerMixin,
  TableMixin,
} from "../mixins";

/**
 * Reusable Role Members Section component.
 * Handles role member list and assignment interactions.
 * Can be composed into any page that displays role members.
 */
export class RoleMembersSection extends DialogMixin(
  SectionContainerMixin(TableMixin(EmptyStateMixin(BaseComponent)))
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='role-members']")
    );
    this.setTableConfig({
      getSectionContainer: () => this.getSectionContainer(),
    });
    this.setEmptyStateConfig({
      emptyStateText: "No members assigned",
      getSectionContainer: () => this.getSectionContainer(),
      getTable: () => this.getTable(),
    });
  }

  /**
   * Override getTable to be more specific - get the table with Name/Email/Actions columns.
   */
  getTable(): Locator {
    return this.getSectionContainer()
      .getByRole("table")
      .filter({ hasText: "Name" })
      .filter({ hasText: "Email" });
  }

  /**
   * Wait for section to load and be visible.
   */
  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
    await this.waitForTableOrEmptyState(options);
  }

  /**
   * Get the add member button locator.
   */
  getAddMemberButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: /add member/i,
    });
  }

  /**
   * Check if the add member button is visible.
   */
  async hasAddMemberButton(): Promise<boolean> {
    const button = this.getAddMemberButton();
    return await button.isVisible().catch(() => false);
  }

  /**
   * Click on the add member button to open the assignment dialog.
   */
  async clickAddMemberButton(): Promise<void> {
    await this.getAddMemberButton().click();
  }

  /**
   * Get a table row by member full name.
   */
  getRowByMemberName(fullName: string): Locator {
    return this.getRowByName(fullName);
  }

  /**
   * Get the remove member button for a specific member.
   */
  getRemoveMemberButton(fullName: string): Locator {
    const row = this.getRowByMemberName(fullName);
    return row.getByRole("button", { name: /remove member/i });
  }

  /**
   * Open the add member dialog and wait for it to be visible.
   */
  async openAddMemberDialog(): Promise<void> {
    await this.clickAddMemberButton();
    await this.waitForDialog("Add Member to Role");
  }

  /**
   * Select a member from the add member dialog combobox.
   *
   * @param memberFullName - The full name of the member to select.
   */
  async selectMemberInDialog(memberFullName: string): Promise<void> {
    const selectTrigger = this.page.getByRole("combobox");
    await selectTrigger.click();

    const memberOption = this.page.getByRole("option", {
      name: memberFullName,
    });
    await memberOption.click();
  }

  /**
   * Confirm the add member dialog.
   */
  async confirmAddMember(): Promise<void> {
    await this.confirmDialog("Add Member");
    await waitForSuccessToast(this.page, "Member added");
  }

  /**
   * Add a member to the role by selecting them from the dialog.
   * This is a convenience method that combines all steps.
   *
   * @param memberFullName - The full name of the member to add.
   */
  async addMember(memberFullName: string): Promise<void> {
    await this.openAddMemberDialog();
    await this.selectMemberInDialog(memberFullName);
    await this.confirmAddMember();
  }

  /**
   * Open the remove member dialog for a specific member.
   *
   * @param memberFullName - The full name of the member to remove.
   */
  async openRemoveMemberDialog(memberFullName: string): Promise<void> {
    const removeButton = this.getRemoveMemberButton(memberFullName);
    await removeButton.click();
    await this.waitForDialog(`Remove ${memberFullName}`);
  }

  /**
   * Confirm the remove member dialog.
   */
  async confirmRemoveMember(): Promise<void> {
    await this.confirmDialog("Remove Member");
    await waitForSuccessToast(this.page, "Member removed");
  }

  /**
   * Remove a member from the role.
   * This is a convenience method that combines all steps.
   *
   * @param memberFullName - The full name of the member to remove.
   */
  async removeMember(memberFullName: string): Promise<void> {
    await this.openRemoveMemberDialog(memberFullName);
    await this.confirmRemoveMember();
  }

  /**
   * Get the count of visible member rows.
   */
  async getMemberCount(): Promise<number> {
    return await this.getRowCount();
  }
}
