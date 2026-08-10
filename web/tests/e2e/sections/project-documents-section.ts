import type { Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { EmptyStateMixin, SectionContainerMixin, TableMixin } from "../mixins";

/**
 * Section abstraction for the project documents list.
 */
export class ProjectDocumentsSection extends SectionContainerMixin(
  TableMixin(EmptyStateMixin(BaseComponent))
) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("div[data-section='project-documents']")
    );
    this.setTableConfig({
      getSectionContainer: () => this.getSectionContainer(),
    });
    this.setEmptyStateConfig({
      emptyStateText: "No documents found",
      getSectionContainer: () => this.getSectionContainer(),
      getTable: () => this.getTable(),
    });
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await this.waitForContainerLoad(options);
    await this.waitForTableOrEmptyState(options);
  }
}
