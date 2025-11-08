import { expect } from "@playwright/test";
import {
  getButtonIn,
  getTableByHeader,
  getTableRow,
  waitForElementVisible,
  waitForLocatorAttached,
  waitForSkeletonToDisappear,
  waitForTableReady,
} from "../helpers/elements";
import { navigateAndWait, waitForPermissionsLoad } from "../helpers/navigation";
import type { Locator, Page } from "@playwright/test";

/**
 * Page Object Model for Role Edit page.
 * Provides methods to interact with role members and permissions.
 */
export class RoleEditPage {
  constructor(
    private page: Page,
    private organizationId: string,
    private roleId: string
  ) {}

  /**
   * Navigate to the role edit page.
   */
  async goto(): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${this.organizationId}/roles/${this.roleId}/edit`
    );
  }

  /**
   * Get the members table locator.
   */
  getMembersTable(): Locator {
    return getTableByHeader(this.page, "Name");
  }

  /**
   * Get a specific member row by member name.
   * @param memberName - Name of the member to find
   */
  getMemberRow(memberName: string): Locator {
    const table = this.getMembersTable();
    return getTableRow(table, memberName);
  }

  /**
   * Wait for members section to load and be visible.
   */
  async waitForMembersLoad(): Promise<void> {
    await waitForElementVisible(
      this.page.getByRole("heading", { name: "Members" }),
      { timeout: 10000 }
    );
    const membersTable = this.getMembersTable();
    await waitForTableReady(membersTable, { timeout: 10000 });
    await waitForPermissionsLoad(this.page);
  }

  /**
   * Click the add member button.
   */
  async clickAddMember(): Promise<void> {
    const addButton = this.page.getByRole("button", { name: /Add Member/i });
    await addButton.click();
  }

  /**
   * Click the remove member button for a specific member.
   * @param memberName - Name of the member to remove
   */
  async clickRemoveMember(memberName: string): Promise<void> {
    const memberRow = this.getMemberRow(memberName);
    const actionsCell = memberRow.locator("td").last();
    await waitForSkeletonToDisappear(actionsCell, { timeout: 15000 });
    await waitForLocatorAttached(actionsCell.locator("button"), {
      timeout: 5000,
    });
    const removeButton = getButtonIn(actionsCell, "Remove member", {
      exact: true,
    });
    await removeButton.waitFor({ state: "visible", timeout: 15000 });
    await removeButton.click();
  }

  /**
   * Check if remove button is visible for a specific member.
   * @param memberName - Name of the member to check
   */
  async hasRemoveButton(memberName: string): Promise<boolean> {
    const memberRow = this.getMemberRow(memberName);
    const actionsCell = memberRow.locator("td").last();
    const removeButton = getButtonIn(actionsCell, "Remove member", {
      exact: true,
    });
    return await removeButton.isVisible().catch(() => false);
  }

  /**
   * Wait for a member to appear in the members table.
   * @param memberName - Name of the member to wait for
   */
  async waitForMember(memberName: string): Promise<void> {
    const memberRow = this.getMemberRow(memberName);
    await waitForElementVisible(memberRow, { timeout: 10000 });
  }

  /**
   * Check if a member exists in the table.
   * @param memberName - Name of the member to check
   */
  async memberExists(memberName: string): Promise<boolean> {
    const memberRow = this.getMemberRow(memberName);
    const count = await memberRow.count().catch(() => 0);
    return count > 0;
  }
}
