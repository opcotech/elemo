import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { waitForSuccessToast } from "../helpers";
import {
  DialogMixin,
  EmptyStateMixin,
  SectionContainerMixin,
  TableMixin,
} from "../mixins";

/**
 * Organization Members section abstraction for the organization details page.
 */
export class OrganizationMembersSection extends DialogMixin(
  SectionContainerMixin(TableMixin(EmptyStateMixin(BaseComponent)))
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='organization-members']")
    );
    this.setTableConfig({
      getSectionContainer: () => this.getSectionContainer(),
    });
    this.setEmptyStateConfig({
      emptyStateText: "No members found",
      getSectionContainer: () => this.getSectionContainer(),
      getTable: () => this.getTable(),
    });
  }

  /**
   * Wait for the members section to be ready before interacting with it.
   */
  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
    await this.waitForTableOrEmptyState(options);
  }

  /**
   * Override table locator to point to the members table within the section.
   */
  getTable(): Locator {
    return this.getSectionContainer().getByRole("table");
  }

  getRowByMemberName(fullName: string): Locator {
    return this.getRowByName(fullName);
  }

  async hasMember(fullName: string): Promise<boolean> {
    return this.hasRow(fullName);
  }

  getInviteMemberButton(): Locator {
    return this.getSectionContainer().getByRole("button", {
      name: /invite member/i,
    });
  }

  async hasInviteMemberButton(): Promise<boolean> {
    const button = this.getInviteMemberButton();
    return await button.isVisible().catch(() => false);
  }

  async clickInviteMemberButton(): Promise<void> {
    await this.getInviteMemberButton().click();
  }

  private getRemoveMemberButton(fullName: string): Locator {
    return this.getRowByMemberName(fullName).getByRole("button", {
      name: /remove member/i,
    });
  }

  async removeMember(fullName: string): Promise<void> {
    await this.openRemoveMemberDialog(fullName);
    await this.confirmRemoveMember();
  }

  async openRemoveMemberDialog(fullName: string): Promise<void> {
    await this.getRemoveMemberButton(fullName).click();
    await this.waitForDialog("Remove Member from Organization?");
  }

  async confirmRemoveMember(): Promise<void> {
    await this.confirmDialog("Remove Member");
    await waitForSuccessToast(this.page, "Member removed");
  }
}
