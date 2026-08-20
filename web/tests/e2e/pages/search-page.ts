import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { navigateAndWait, waitForElementVisible } from "../helpers";

/**
 * Page Object Model for /search.
 */
export class SearchPage extends BaseComponent {
  constructor(page: Page) {
    super(page);
  }

  async goto(search?: {
    q?: string;
    type?: string;
    organization_id?: string;
    namespace_id?: string;
    project_id?: string;
  }): Promise<void> {
    const params = new URLSearchParams();
    if (search?.q) params.set("q", search.q);
    if (search?.type) params.set("type", search.type);
    if (search?.organization_id) {
      params.set("organization_id", search.organization_id);
    }
    if (search?.namespace_id) params.set("namespace_id", search.namespace_id);
    if (search?.project_id) params.set("project_id", search.project_id);
    const query = params.toString();
    await navigateAndWait(this.page, query ? `/search?${query}` : "/search", {
      ready: this.getHeading(),
    });
  }

  async waitForLoad(options?: { timeout?: number }): Promise<void> {
    await waitForElementVisible(this.getHeading(), options);
  }

  getHeading(): Locator {
    return this.page.getByRole("heading", { name: "Search" });
  }

  getQueryInput(): Locator {
    return this.page.getByRole("searchbox", { name: "Search" });
  }

  getTypeFilter(): Locator {
    return this.page.getByRole("combobox", { name: "Filter by type" });
  }

  getNamespaceFilter(): Locator {
    return this.page.getByRole("combobox", { name: "Filter by namespace" });
  }

  getResultLink(title: string): Locator {
    return this.page.getByRole("link", { name: title }).first();
  }
}
