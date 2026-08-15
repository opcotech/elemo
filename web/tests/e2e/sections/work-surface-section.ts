import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import {
  fillLocator,
  waitForAnimations,
  waitForElementVisible,
} from "../helpers";
import { SectionContainerMixin } from "../mixins";

export type WorkLayoutName = "List" | "Table" | "Board" | "Timeline";
export type WorkGroupName = "Status" | "Priority" | "Assignee" | "No grouping";
export type WorkSortName =
  "Manual rank" | "Priority" | "Due date" | "Recently updated";
export type WorkDisplayName = "Comfortable" | "Compact";

/**
 * Work surface toolbar, layouts, and inspect actions.
 */
export class WorkSurfaceSection extends SectionContainerMixin(BaseComponent) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='work-surface']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  getCreateButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Create",
      exact: true,
    });
  }

  async clickCreate(): Promise<void> {
    await this.getCreateButton().click();
  }

  getQuickCreateButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Quick create",
    });
  }

  async clickQuickCreate(): Promise<void> {
    await this.getQuickCreateButton().click();
  }

  getLayoutButton(layout: WorkLayoutName): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: layout,
      exact: true,
    });
  }

  async selectLayout(layout: WorkLayoutName): Promise<void> {
    await this.getLayoutButton(layout).click();
  }

  getFilterButton(): Locator {
    return this.getSectionContainer().getByRole("button", { name: "Filter" });
  }

  getFilterInput(): Locator {
    return this.page.getByLabel("Filter work");
  }

  async openFilter(): Promise<void> {
    await this.getFilterButton().click();
    await waitForElementVisible(this.getFilterInput());
    await waitForAnimations(this.getFilterInput());
  }

  async fillFilter(value: string): Promise<void> {
    await this.openFilter();
    await fillLocator(this.getFilterInput(), value);
  }

  async clearFilter(): Promise<void> {
    const chip = this.getSectionContainer().getByRole("button", {
      name: /×$/,
    });
    await chip.click();
  }

  getGroupButton(): Locator {
    return this.getSectionContainer().getByRole("button", { name: "Group by" });
  }

  async selectGroup(name: WorkGroupName): Promise<void> {
    await this.getGroupButton().click();
    await this.page.getByRole("menuitemradio", { name }).click();
  }

  getSortButton(): Locator {
    return this.getSectionContainer().getByRole("button", { name: "Sort" });
  }

  async selectSort(name: WorkSortName): Promise<void> {
    await this.getSortButton().click();
    await this.page.getByRole("menuitemradio", { name }).click();
  }

  getDisplayButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Display density",
    });
  }

  async selectDisplay(name: WorkDisplayName): Promise<void> {
    await this.getDisplayButton().click();
    await this.page.getByRole("menuitemradio", { name }).click();
  }

  getInspectButton(key: string, title?: string): Locator {
    const name = title
      ? new RegExp(`^Inspect ${key}: ${title}`)
      : new RegExp(`^Inspect ${key}:`);
    return this.getSectionContainer().getByRole("button", { name });
  }

  async inspect(key: string, title?: string): Promise<void> {
    await this.getInspectButton(key, title).first().click();
  }

  getWorkKeyLink(key: string): Locator {
    return this.getSectionContainer().getByRole("link", {
      name: key,
      exact: true,
    });
  }

  getAddWorkToColumnButton(column: string): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: `Add work to ${column}`,
    });
  }

  getEmptyState(): Locator {
    return this.getSectionContainer().getByText("No work yet");
  }
}
