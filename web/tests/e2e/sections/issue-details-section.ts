import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { clickUntilVisible, waitForDropdownOpen } from "../helpers";
import { SectionContainerMixin } from "../mixins";

/**
 * Issue details sidebar: kind, status, people, dates, and parent.
 */
export class IssueDetailsSection extends SectionContainerMixin(BaseComponent) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='issue-details']")
    );
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
  }

  private getMetadataContainer(): Locator {
    return this.page.locator("[data-section='issue-metadata']");
  }

  getKindSelect(): Locator {
    return this.getSectionContainer().getByLabel("Kind", { exact: true });
  }

  getStatusSelect(): Locator {
    return this.getSectionContainer().getByLabel("Status", { exact: true });
  }

  getResolutionSelect(): Locator {
    return this.getSectionContainer().getByLabel("Resolution", { exact: true });
  }

  getPrioritySelect(): Locator {
    return this.getSectionContainer().getByLabel("Priority", { exact: true });
  }

  getAssigneesSelect(): Locator {
    return this.getSectionContainer().getByLabel("Assignees", { exact: true });
  }

  getReviewersSelect(): Locator {
    return this.getSectionContainer().getByLabel("Reviewers", { exact: true });
  }

  getStartDatePicker(): Locator {
    return this.getSectionContainer().getByLabel("Start date", { exact: true });
  }

  getDueDatePicker(): Locator {
    return this.getSectionContainer().getByLabel("Due date", { exact: true });
  }

  getClearStartDateButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Clear start date",
    });
  }

  getClearDueDateButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: "Clear due date",
    });
  }

  getParentSelect(): Locator {
    return this.getMetadataContainer().getByLabel("Parent", { exact: true });
  }

  async selectKind(name: string): Promise<void> {
    await this.selectOption(this.getKindSelect(), name);
  }

  async selectStatus(name: string): Promise<void> {
    await this.selectOption(this.getStatusSelect(), name);
  }

  async selectResolution(name: string): Promise<void> {
    await this.selectOption(this.getResolutionSelect(), name);
  }

  async selectPriority(name: string): Promise<void> {
    await this.selectOption(this.getPrioritySelect(), name);
  }

  async openAssignees(): Promise<void> {
    await this.getAssigneesSelect().click();
  }

  async openReviewers(): Promise<void> {
    await this.getReviewersSelect().click();
  }

  async selectPersonOption(name: string): Promise<void> {
    await this.page
      .locator('[data-slot="command-item"]')
      .filter({ hasText: name })
      .first()
      .click();
  }

  async selectParent(name: string): Promise<void> {
    await this.getParentSelect().click();
    await this.page
      .locator('[data-slot="command-item"]')
      .filter({ hasText: name })
      .first()
      .click();
  }

  async clearParent(): Promise<void> {
    await this.selectParent("None");
  }

  async openStartDatePicker(): Promise<void> {
    await clickUntilVisible(
      this.getStartDatePicker(),
      this.page.locator("[data-slot='calendar']"),
      { timeout: 8_000 }
    );
  }

  async openDueDatePicker(): Promise<void> {
    await clickUntilVisible(
      this.getDueDatePicker(),
      this.page.locator("[data-slot='calendar']"),
      { timeout: 8_000 }
    );
  }

  private async selectOption(trigger: Locator, name: string): Promise<void> {
    await trigger.scrollIntoViewIfNeeded();
    await trigger.click();
    try {
      await waitForDropdownOpen(trigger, { timeout: 2_000 });
    } catch {
      await trigger.click();
      await waitForDropdownOpen(trigger);
    }
    await this.page.getByRole("option", { name, exact: true }).click();
  }
}
