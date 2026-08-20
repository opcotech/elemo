import { expect } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { SectionContainerMixin } from "../mixins";

import type { Action } from "@/lib/api/types";

/**
 * Section for managing bundled actions on the role edit page.
 * Prefer API helpers when the actions UI is not yet available.
 */
export class RolePermissionsSection extends SectionContainerMixin(
  BaseComponent
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

  getActionRow(action: Action): Locator {
    return this.getSectionContainer()
      .getByRole("row")
      .filter({ hasText: action });
  }

  async waitForActionRow(action: Action): Promise<Locator> {
    const row = this.getActionRow(action).first();
    await expect(row).toBeVisible();
    return row;
  }
}
