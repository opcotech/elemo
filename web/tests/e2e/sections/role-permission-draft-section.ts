import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { SectionContainerMixin } from "../mixins";

import type { PermissionKind } from "@/lib/api";

interface DraftPermissionOptions {
  resourceType: string;
  resourceId: string;
  permissionKind: PermissionKind;
}

/**
 * Section for managing the pending permissions draft on the role create page.
 */
export class RolePermissionDraftSection extends SectionContainerMixin(
  BaseComponent
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='role-permission-draft']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  private getPermissionForm(): Locator {
    return this.getSectionContainer().locator("form").first();
  }

  private getPermissionTable(): Locator {
    return this.getSectionContainer().getByRole("table");
  }

  getPermissionRow(
    resourceId: string,
    permissionKind: PermissionKind
  ): Locator {
    return this.getPermissionTable()
      .getByRole("row")
      .filter({ hasText: resourceId })
      .filter({ hasText: permissionKind });
  }

  async waitForPermissionRow(
    resourceId: string,
    permissionKind: PermissionKind
  ): Promise<Locator> {
    const row = this.getPermissionRow(resourceId, permissionKind).first();
    await expect(row).toBeVisible();
    return row;
  }

  async addPermission({
    resourceType,
    resourceId,
    permissionKind,
  }: DraftPermissionOptions): Promise<void> {
    const form = this.getPermissionForm();
    const comboboxes = form.getByRole("combobox");

    await comboboxes.first().click();
    await this.page
      .getByRole("option", { name: new RegExp(resourceType, "i") })
      .click();

    await form.getByPlaceholder("Enter resource ID").fill(resourceId);

    await comboboxes.nth(1).click();
    await this.page
      .getByRole("option", { name: new RegExp(permissionKind, "i") })
      .click();

    await form.getByRole("button", { name: "Add Permission" }).click();

    await this.waitForPermissionRow(resourceId, permissionKind);
  }
}
