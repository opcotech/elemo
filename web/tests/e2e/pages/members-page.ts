import {
  getButtonIn,
  getTableByHeader,
  getTableRow,
  waitForElementVisible,
  waitForTableReady,
} from "../helpers/elements";
import { navigateAndWait, waitForPermissionsLoad } from "../helpers/navigation";
import type { Locator, Page } from "@playwright/test";

/**
 * Page Object Model for Organization Members page.
 * Provides methods to interact with organization members section.
 */
export class MembersPage {
  constructor(
    private page: Page,
    private organizationId: string
  ) {}

  /**
   * Navigate to the organization page (members section is on the same page).
   */
  async goto(): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${this.organizationId}`
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
   * @param memberName - Name of the member to find (e.g., "John Doe")
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
      this.page.getByRole("heading", { name: "Members", level: 3 }),
      { timeout: 10000 }
    );
    const membersTable = this.getMembersTable();
    await waitForTableReady(membersTable, { timeout: 10000 });
    await waitForPermissionsLoad(this.page, this.organizationId);
  }

  /**
   * Click the invite member button.
   */
  async clickInviteMember(): Promise<void> {
    const inviteButton = this.page.getByRole("button", {
      name: /Invite Member/i,
    });
    await inviteButton.click();
  }

  /**
   * Click the remove member button for a specific member.
   * @param memberName - Name of the member to remove
   */
  async clickRemoveMember(memberName: string): Promise<void> {
    const memberRow = this.getMemberRow(memberName);
    const actionsCell = memberRow.locator("td").last();
    const removeButton = getButtonIn(actionsCell, "Remove member", {
      exact: true,
    });
    await removeButton.click();
  }

  /**
   * Check if invite button is visible.
   */
  async hasInviteButton(): Promise<boolean> {
    const inviteButton = this.page.getByRole("button", {
      name: /Invite Member/i,
    });
    return await inviteButton.isVisible().catch(() => false);
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
    return (await memberRow.count()) > 0;
  }

  /**
   * Get the count of members in the table.
   */
  async getMemberCount(): Promise<number> {
    const table = this.getMembersTable();
    const rows = table.locator("tbody tr");
    return await rows.count();
  }
}
