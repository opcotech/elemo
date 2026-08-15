import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import {
  clickUntilVisible,
  fillLocator,
  navigateAndWait,
  waitForElementVisible,
} from "../helpers";
import { IssueDetailsSection } from "../sections/issue-details-section";
import { IssueLinksSection } from "../sections/issue-links-section";
import { IssueRelationsSection } from "../sections/issue-relations-section";

/**
 * Page Object Model for a work item at /work/:namespaceId/:issueKey.
 */
export class WorkItemPage extends BaseComponent {
  public readonly details: IssueDetailsSection;
  public readonly links: IssueLinksSection;
  public readonly relations: IssueRelationsSection;

  constructor(page: Page) {
    super(page);
    this.details = new IssueDetailsSection(page);
    this.links = new IssueLinksSection(page);
    this.relations = new IssueRelationsSection(page);
  }

  async goto(namespaceId: string, issueKey: string): Promise<void> {
    await navigateAndWait(this.page, `/work/${namespaceId}/${issueKey}`);
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await waitForElementVisible(
      this.page.locator("[data-section='issue-details']"),
      options
    );
  }

  getCopyLinkButton(): Locator {
    return this.page.getByRole("button", { name: "Copy issue link" });
  }

  async copyLink(): Promise<void> {
    await this.getCopyLinkButton().click();
  }

  getMoreActionsButton(): Locator {
    return this.page.getByRole("button", { name: "More actions" });
  }

  async openMoreActions(): Promise<void> {
    await clickUntilVisible(
      this.getMoreActionsButton(),
      this.page.getByRole("menu")
    );
  }

  async clickMoreAction(name: string): Promise<void> {
    await this.openMoreActions();
    await this.page.getByRole("menuitem", { name }).click();
  }

  getTitleButton(): Locator {
    return this.page.getByRole("button", { name: "Edit title" });
  }

  getTitleInput(): Locator {
    return this.page.getByLabel("Issue title");
  }

  async editTitle(title: string): Promise<void> {
    await this.getTitleButton().click();
    await fillLocator(this.getTitleInput(), title);
    await this.getTitleInput().press("Enter");
  }

  getDescriptionSection(): Locator {
    return this.page.locator("[data-section='issue-description']");
  }

  getEditDescriptionButton(): Locator {
    return this.getDescriptionSection().getByRole("button", {
      name: "Edit description",
    });
  }

  async startDescriptionEdit(): Promise<void> {
    await this.getEditDescriptionButton().click();
  }

  async saveDescription(): Promise<void> {
    await this.getDescriptionSection()
      .getByRole("button", { name: "Save" })
      .click();
  }

  async cancelDescriptionEdit(): Promise<void> {
    await this.getDescriptionSection()
      .getByRole("button", { name: "Cancel" })
      .click();
  }

  async openDeleteDialog(): Promise<void> {
    await this.clickMoreAction("Delete");
  }

  async confirmDelete(): Promise<void> {
    await this.page.getByRole("button", { name: "Delete issue" }).click();
  }
}
