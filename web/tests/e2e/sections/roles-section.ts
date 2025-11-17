import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import {
  getElementByText,
  waitForPermissionsLoad,
  waitForSuccessToast,
} from "../helpers";
import {
  DialogMixin,
  EmptyStateMixin,
  SearchMixin,
  SectionContainerMixin,
  TableMixin,
} from "../mixins";

/**
 * Reusable Roles Section component.
 * Handles roles table interactions.
 * Can be composed into any page that displays roles.
 */
export class RolesSection extends DialogMixin(
  SectionContainerMixin(TableMixin(SearchMixin(EmptyStateMixin(BaseComponent))))
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(this.page.locator("[data-section='roles']"));
    this.setTableConfig({
      getSectionContainer: () => this.getSectionContainer(),
    });
    this.setSearchConfig({
      getSectionContainer: () => this.getSectionContainer(),
      searchPlaceholder: "Search roles...",
    });
    this.setEmptyStateConfig({
      emptyStateText: "No roles found",
      getSectionContainer: () => this.getSectionContainer(),
      getTable: () => this.getTable(),
    });
  }

  /**
   * Wait for section to load and be visible.
   */
  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
    await this.waitForTableOrEmptyState(options);
    await waitForPermissionsLoad(this.page);
  }

  /**
   * Get a table row by role name.
   */
  getRowByRoleName(name: string): Locator {
    return this.getRowByName(name);
  }

  /**
   * Check if a role row exists in the table.
   */
  async hasRole(name: string): Promise<boolean> {
    return await this.hasRow(name);
  }

  /**
   * Get the count of visible role rows.
   */
  async getRoleCount(): Promise<number> {
    return await this.getRowCount();
  }

  /**
   * Get the create role button locator.
   */
  getCreateRoleButton(): Locator {
    return getElementByText(this.getSectionContainer(), "Create Role");
  }

  /**
   * Check if the create role button is visible.
   */
  async hasCreateRoleButton(): Promise<boolean> {
    const button = this.getCreateRoleButton();
    return await button.isVisible().catch(() => false);
  }

  /**
   * Click on the create role button.
   */
  async clickCreateRoleButton(): Promise<void> {
    await this.getCreateRoleButton().click();
  }

  /**
   * Get the delete button for a specific role.
   */
  getDeleteRoleButton(roleName: string): Locator {
    const row = this.getRowByRoleName(roleName);
    return row.getByRole("button", { name: /delete role/i });
  }

  /**
   * Check if the delete button is visible for a specific role.
   */
  async hasDeleteRoleButton(roleName: string): Promise<boolean> {
    const button = this.getDeleteRoleButton(roleName);
    return await button.isVisible().catch(() => false);
  }

  /**
   * Open the delete role dialog for a specific role.
   *
   * @param roleName - The name of the role to delete.
   */
  async openDeleteRoleDialog(roleName: string): Promise<void> {
    const deleteButton = this.getDeleteRoleButton(roleName);
    await deleteButton.click();
    await this.waitForDialog(`Are you sure you want to delete ${roleName}?`);
  }

  /**
   * Confirm the delete role dialog.
   */
  async confirmDeleteRole(): Promise<void> {
    await this.confirmDialog("Delete");
    await waitForSuccessToast(this.page, "deleted");
  }

  /**
   * Delete a role.
   * This is a convenience method that combines all steps.
   *
   * @param roleName - The name of the role to delete.
   */
  async deleteRole(roleName: string): Promise<void> {
    await this.openDeleteRoleDialog(roleName);
    await this.confirmDeleteRole();
  }
}
