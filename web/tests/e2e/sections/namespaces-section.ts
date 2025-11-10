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
 * Section abstraction for the organization namespaces list.
 */
export class NamespacesSection extends DialogMixin(
  SectionContainerMixin(TableMixin(EmptyStateMixin(BaseComponent)))
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='organization-namespaces']")
    );
    this.setTableConfig({
      getSectionContainer: () => this.getSectionContainer(),
    });
    this.setEmptyStateConfig({
      emptyStateText: "No namespaces found",
      getSectionContainer: () => this.getSectionContainer(),
      getTable: () => this.getTable(),
    });
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
    await this.waitForTableOrEmptyState(options);
    await waitForPermissionsLoad(this.page);
  }

  getRowByNamespaceName(name: string): Locator {
    return this.getRowByName(name);
  }

  async hasNamespace(name: string): Promise<boolean> {
    return this.hasRow(name);
  }

  private getCreateNamespaceButtonLocator(): Locator {
    const container = this.getSectionContainer();
    return container
      .getByRole("button", { name: /create namespace/i })
      .or(container.getByRole("link", { name: /create namespace/i }));
  }

  async hasCreateNamespaceButton(): Promise<boolean> {
    return (await this.getCreateNamespaceButtonLocator().count()) > 0;
  }

  async clickCreateNamespaceButton(): Promise<void> {
    await this.getCreateNamespaceButtonLocator().first().click();
  }

  private getDeleteNamespaceButton(name: string): Locator {
    return this.getRowByNamespaceName(name).getByRole("button", {
      name: /delete namespace/i,
    });
  }

  async hasDeleteNamespaceButton(name: string): Promise<boolean> {
    return await this.getDeleteNamespaceButton(name)
      .isVisible()
      .catch(() => false);
  }

  async deleteNamespace(name: string): Promise<void> {
    await this.openDeleteNamespaceDialog(name);
    await this.confirmDeleteNamespace();
  }

  async openDeleteNamespaceDialog(name: string): Promise<void> {
    await this.getDeleteNamespaceButton(name).click();
    await this.waitForDialog(`Are you sure you want to delete ${name}?`);
  }

  async confirmDeleteNamespace(): Promise<void> {
    await this.confirmDialog("Delete");
    await waitForSuccessToast(this.page, "Namespace deleted");
  }
}
