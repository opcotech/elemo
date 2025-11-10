import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { getElementByText, waitForPermissionsLoad } from "../helpers";
import {
  EmptyStateMixin,
  SearchMixin,
  SectionContainerMixin,
  TableMixin,
} from "../mixins";

/**
 * Reusable Organizations Section component.
 * Handles organizations table interactions.
 * Can be composed into any page that displays organizations.
 */
export class OrganizationsSection extends SectionContainerMixin(
  TableMixin(SearchMixin(EmptyStateMixin(BaseComponent)))
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='organizations']")
    );
    this.setTableConfig({
      getSectionContainer: () => this.getSectionContainer(),
    });
    this.setSearchConfig({
      getSectionContainer: () => this.getSectionContainer(),
      searchPlaceholder: "Search organizations...",
    });
    this.setEmptyStateConfig({
      emptyStateText: "No organizations available",
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
   * Get a table row by organization name.
   */
  getRowByOrganizationName(name: string): Locator {
    return this.getRowByName(name);
  }

  /**
   * Get the organization name link locator.
   */
  getOrganizationLink(name: string): Locator {
    return this.getLinkByName(name);
  }

  /**
   * Click on an organization name to navigate to details.
   */
  async clickOrganization(name: string): Promise<void> {
    await this.clickLink(name);
  }

  /**
   * Check if an organization row exists in the table.
   */
  async hasOrganization(name: string): Promise<boolean> {
    return await this.hasRow(name);
  }

  /**
   * Get the count of visible organization rows.
   */
  async getOrganizationCount(): Promise<number> {
    return await this.getRowCount();
  }

  /**
   * Get the create organization button locator.
   */
  getCreateOrganizationButton(): Locator {
    return getElementByText(this.getSectionContainer(), "Create Organization");
  }

  /**
   * Check if the create organization button is visible.
   */
  async hasCreateOrganizationButton(): Promise<boolean> {
    const button = this.getCreateOrganizationButton();
    return await button.isVisible().catch(() => false);
  }

  /**
   * Click on the create organization button.
   */
  async clickCreateOrganizationButton(): Promise<void> {
    await this.getCreateOrganizationButton().click();
  }
}
