import { expect } from "@playwright/test";
import {
  getButtonIn,
  getTableRow,
  waitForElementVisible,
  waitForLocatorAttached,
  waitForSkeletonToDisappear,
  waitForTableReady,
} from "../helpers/elements";
import { navigateAndWait, waitForPermissionsLoad } from "../helpers/navigation";
import type { Locator, Page } from "@playwright/test";

/**
 * Page Object Model for Organization settings page.
 * Provides methods to interact with organization details, roles, and members.
 */
export class OrganizationPage {
  constructor(
    private page: Page,
    private organizationId: string
  ) {}

  /**
   * Navigate to the organization page.
   */
  async goto(): Promise<void> {
    await navigateAndWait(
      this.page,
      `/settings/organizations/${this.organizationId}`
    );
  }

  /**
   * Get the roles table locator.
   */
  getRolesTable(): Locator {
    return this.page
      .locator("table")
      .filter({
        has: this.page.locator("thead th", { hasText: "Description" }),
      })
      .first();
  }

  /**
   * Get a specific role row by role name.
   * @param roleName - Name of the role to find
   */
  getRoleRow(roleName: string): Locator {
    const table = this.getRolesTable();
    return getTableRow(table, roleName);
  }

  /**
   * Wait for roles section to load and be visible.
   */
  async waitForRolesLoad(): Promise<void> {
    await waitForElementVisible(
      this.page.getByRole("heading", { name: "Roles" }),
      { timeout: 10000 }
    );
    const rolesTable = this.getRolesTable();
    await waitForTableReady(rolesTable, { timeout: 10000 });
    await waitForPermissionsLoad(this.page);
    const skeletons = this.page.locator('[data-slot="skeleton"]');
    try {
      await skeletons
        .first()
        .waitFor({ state: "hidden", timeout: 15000 })
        .catch(() => {});
      await this.page.waitForTimeout(500);
    } catch {}
  }

  /**
   * Click the create role button.
   */
  async clickCreateRole(): Promise<void> {
    const createButton = this.page
      .getByRole("link", { name: /Create Role/i })
      .first();
    await createButton.click();
  }

  /**
   * Click the edit button for a specific role.
   * @param roleName - Name of the role to edit
   */
  async clickEditRole(roleName: string): Promise<void> {
    const roleRow = this.getRoleRow(roleName);
    const editButton = roleRow.getByRole("link", { name: "Edit role" });
    await editButton.click();
  }

  /**
   * Click the delete button for a specific role.
   * @param roleName - Name of the role to delete
   */
  async clickDeleteRole(roleName: string): Promise<void> {
    const roleRow = this.getRoleRow(roleName);
    const actionsCell = roleRow.locator("td").last();
    await waitForSkeletonToDisappear(actionsCell, { timeout: 15000 });
    await waitForLocatorAttached(actionsCell.locator("button"), {
      timeout: 5000,
    });

    const deleteButton = getButtonIn(actionsCell, "Delete role");
    if ((await deleteButton.count()) === 0) {
      throw new Error("Delete role button not found");
    }

    await deleteButton.waitFor({ state: "visible", timeout: 5000 });
    await deleteButton.click();
  }

  /**
   * Check if delete button is visible for a specific role.
   * @param roleName - Name of the role to check
   */
  async hasDeleteButton(roleName: string): Promise<boolean> {
    const roleRow = this.getRoleRow(roleName);
    const actionsCell = roleRow.locator("td").last();

    await waitForSkeletonToDisappear(actionsCell, { timeout: 3000 });
    await waitForLocatorAttached(actionsCell.locator("button"), {
      timeout: 5000,
    });

    const deleteButton = getButtonIn(actionsCell, "Delete role");
    if ((await deleteButton.count()) === 0) {
      return false;
    }

    try {
      await deleteButton.waitFor({ state: "visible", timeout: 2000 });
      return await deleteButton.isVisible();
    } catch {
      return false;
    }
  }

  /**
   * Wait for a role to appear in the roles table.
   * @param roleName - Name of the role to wait for
   */
  async waitForRole(roleName: string): Promise<void> {
    const roleRow = this.getRoleRow(roleName);
    await waitForElementVisible(roleRow, { timeout: 10000 });
  }

  /**
   * Check if a role exists in the table.
   * @param roleName - Name of the role to check
   */
  async roleExists(roleName: string): Promise<boolean> {
    const roleRow = this.getRoleRow(roleName);
    return (await roleRow.count()) > 0;
  }

  /**
   * Get the members table locator.
   */
  getMembersTable(): Locator {
    return this.page
      .locator("table")
      .filter({
        has: this.page.locator("thead th", { hasText: "Status" }),
      })
      .first();
  }

  /**
   * Get a specific member row by member name or email.
   * @param memberIdentifier - Name or email of the member to find
   */
  getMemberRow(memberIdentifier: string): Locator {
    const table = this.getMembersTable();
    return getTableRow(table, memberIdentifier);
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
    await waitForElementVisible(membersTable, { timeout: 10000 });
    await waitForPermissionsLoad(this.page);
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
   * @param memberIdentifier - Name or email of the member to remove
   */
  async clickRemoveMember(memberIdentifier: string): Promise<void> {
    const memberRow = this.getMemberRow(memberIdentifier);
    const removeButton = memberRow.getByRole("button", {
      name: "Remove member",
    });
    await removeButton.click();
  }

  /**
   * Check if remove button is visible for a specific member.
   * @param memberIdentifier - Name or email of the member to check
   */
  async hasRemoveButton(memberIdentifier: string): Promise<boolean> {
    const memberRow = this.getMemberRow(memberIdentifier);
    const removeButton = memberRow.getByRole("button", {
      name: "Remove member",
    });
    return await removeButton.isVisible().catch(() => false);
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
   * Wait for a member to appear in the members table.
   * @param memberIdentifier - Name or email of the member to wait for
   */
  async waitForMember(memberIdentifier: string): Promise<void> {
    const memberRow = this.getMemberRow(memberIdentifier);
    await waitForElementVisible(memberRow, { timeout: 10000 });
  }

  /**
   * Check if a member exists in the table.
   * @param memberIdentifier - Name or email of the member to check
   */
  async memberExists(memberIdentifier: string): Promise<boolean> {
    const memberRow = this.getMemberRow(memberIdentifier);
    return (await memberRow.count()) > 0;
  }
}
