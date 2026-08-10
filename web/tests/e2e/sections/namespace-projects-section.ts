import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { waitForPermissionsLoad, waitForSuccessToast } from "../helpers";
import {
  DialogMixin,
  EmptyStateMixin,
  SectionContainerMixin,
  TableMixin,
} from "../mixins";

/**
 * Section abstraction for the namespace projects list.
 */
export class NamespaceProjectsSection extends DialogMixin(
  SectionContainerMixin(TableMixin(EmptyStateMixin(BaseComponent)))
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='namespace-projects']")
    );
    this.setTableConfig({
      getSectionContainer: () => this.getSectionContainer(),
    });
    this.setEmptyStateConfig({
      emptyStateText: "No projects found",
      getSectionContainer: () => this.getSectionContainer(),
      getTable: () => this.getTable(),
    });
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
    await this.waitForTableOrEmptyState(options);
    await waitForPermissionsLoad(this.page);
    // Per-row project permission checks render skeletons while loading.
    await this.getSectionContainer()
      .locator(".animate-pulse")
      .first()
      .waitFor({ state: "hidden", timeout: options?.timeout ?? 10000 })
      .catch(() => undefined);
  }

  getRowByProjectName(name: string): Locator {
    return this.getRowByName(name);
  }

  async hasProject(name: string): Promise<boolean> {
    return this.hasRow(name);
  }

  private getCreateProjectButtonLocator(): Locator {
    const container = this.getSectionContainer();
    return container
      .getByRole("button", { name: /create project/i })
      .or(container.getByRole("link", { name: /create project/i }));
  }

  async hasCreateProjectButton(): Promise<boolean> {
    return (await this.getCreateProjectButtonLocator().count()) > 0;
  }

  async clickCreateProjectButton(): Promise<void> {
    await this.getCreateProjectButtonLocator().first().click();
  }

  async clickProjectLink(name: string): Promise<void> {
    await this.clickLink(name);
  }

  private getEditProjectLink(name: string): Locator {
    return this.getRowByProjectName(name)
      .getByRole("link")
      .filter({ hasText: /edit project/i });
  }

  async hasEditProjectButton(name: string): Promise<boolean> {
    return await this.getEditProjectLink(name)
      .isVisible()
      .catch(() => false);
  }

  async clickEditProjectButton(name: string): Promise<void> {
    await this.getEditProjectLink(name).click();
  }

  private getDeleteProjectButton(name: string): Locator {
    return this.getRowByProjectName(name).getByRole("button", {
      name: /delete project/i,
    });
  }

  async hasDeleteProjectButton(name: string): Promise<boolean> {
    return await this.getDeleteProjectButton(name)
      .isVisible()
      .catch(() => false);
  }

  async openDeleteProjectDialog(name: string): Promise<void> {
    await this.getDeleteProjectButton(name).click();
    await this.waitForDialog(`Are you sure you want to delete ${name}?`);
  }

  async confirmDeleteProject(): Promise<void> {
    await this.confirmDialog("Delete");
    await waitForSuccessToast(this.page, "Project deleted");
  }
}
