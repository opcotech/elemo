import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { fillLocator, waitForSuccessToast } from "../helpers";
import { DialogMixin, SectionContainerMixin } from "../mixins";

import type { PermissionKind } from "@/lib/api/types";

interface PermissionOptions {
  resourceType: string;
  resourceId: string;
  permissionKind: PermissionKind;
}

/**
 * Section for managing assigned permissions on the role edit page.
 */
export class RolePermissionsSection extends DialogMixin(
  SectionContainerMixin(BaseComponent)
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='role-permissions']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  private getPermissionsTable(): Locator {
    return this.getSectionContainer().getByRole("table");
  }

  getPermissionRow(
    resourceType: string,
    resourceId: string,
    permissionKind: PermissionKind
  ): Locator {
    return this.getPermissionsTable()
      .getByRole("row")
      .filter({ hasText: resourceType })
      .filter({ hasText: resourceId })
      .filter({ hasText: permissionKind });
  }

  async waitForPermissionRow(
    resourceType: string,
    resourceId: string,
    permissionKind: PermissionKind
  ): Promise<Locator> {
    const row = this.getPermissionRow(
      resourceType,
      resourceId,
      permissionKind
    ).first();
    await expect(row).toBeVisible();
    return row;
  }

  async waitForPermissionRemoval(
    resourceType: string,
    resourceId: string,
    permissionKind: PermissionKind
  ): Promise<void> {
    await expect(
      this.getPermissionRow(resourceType, resourceId, permissionKind)
    ).not.toBeVisible();
  }

  private getAddPermissionButton(): Locator {
    return this.getSectionContainer()
      .locator("[data-slot='card-header']")
      .getByRole("button", {
        name: /^Add Permission$/i,
      })
      .first();
  }

  async openAddPermissionDialog(): Promise<Locator> {
    const addButton = this.getAddPermissionButton();
    await expect(addButton).toBeVisible();
    await expect(addButton).toBeEnabled();

    for (let attempt = 0; attempt < 2; attempt++) {
      await addButton.click();
      try {
        await this.waitForDialog("Add Permission", { timeout: 5_000 });
        break;
      } catch (error) {
        if (attempt === 1) {
          throw error;
        }
      }
    }

    return this.page
      .getByRole("dialog", { name: "Add Permission" })
      .or(
        this.page.locator('[data-slot="dialog-content"]').filter({
          hasText: "Add Permission",
        })
      )
      .first();
  }

  async addPermission({
    resourceType,
    resourceId,
    permissionKind,
  }: PermissionOptions): Promise<void> {
    const dialog = await this.openAddPermissionDialog();
    const comboboxes = dialog.getByRole("combobox");

    await comboboxes.first().click();
    await this.page
      .getByRole("option", { name: new RegExp(resourceType, "i") })
      .click();

    await fillLocator(dialog.getByPlaceholder("Enter resource ID"), resourceId);

    await comboboxes.nth(1).click();
    await this.page
      .getByRole("option", { name: new RegExp(permissionKind, "i") })
      .click();

    await dialog.getByRole("button", { name: /^Add Permission$/ }).click();
    await waitForSuccessToast(this.page, "Permission added");
    await this.waitForPermissionRow(resourceType, resourceId, permissionKind);
  }

  async removePermission({
    resourceType,
    resourceId,
    permissionKind,
  }: PermissionOptions): Promise<void> {
    const row = await this.waitForPermissionRow(
      resourceType,
      resourceId,
      permissionKind
    );

    await row.getByRole("button", { name: /delete permission/i }).click();

    await this.waitForDialog("Remove Permission?");
    const deleteDialog = this.page.getByRole("alertdialog", {
      name: /Remove Permission/i,
    });

    await deleteDialog
      .getByRole("button", { name: "Remove Permission" })
      .click();

    await waitForSuccessToast(this.page, "Permission removed");
    await this.waitForPermissionRemoval(
      resourceType,
      resourceId,
      permissionKind
    );
  }
}
