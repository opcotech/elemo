import type { Locator, Page } from "@playwright/test";

import { BaseComponent } from "../components/base";
import { waitForElementVisible } from "../helpers";
import { SectionContainerMixin } from "../mixins";

/**
 * Work inspector overlay (`{KEY} details`).
 */
export class WorkInspectorSection extends SectionContainerMixin(BaseComponent) {
  constructor(page: Page) {
    super(page);
    this.setSectionContainer(
      this.page.locator("[data-section='work-inspector']")
    );
  }

  getInspector(key?: string): Locator {
    if (key) {
      return this.page.getByRole("complementary", { name: `${key} details` });
    }
    return this.getSectionContainer();
  }

  async waitForLoad(
    key?: string,
    options?: { timeout?: number }
  ): Promise<void> {
    await waitForElementVisible(this.getInspector(key), options);
  }

  getCloseButton(key?: string): Locator {
    return this.getInspector(key).getByRole("button", {
      name: "Close inspector",
    });
  }

  async close(key?: string): Promise<void> {
    await this.getCloseButton(key).click();
  }

  getOpenFullPageButton(key?: string): Locator {
    const inspector = this.getInspector(key);
    return inspector
      .getByRole("link", { name: "Open full page" })
      .or(inspector.getByRole("button", { name: "Open full page" }));
  }

  async openFullPage(key?: string): Promise<void> {
    await this.getOpenFullPageButton(key).click();
  }

  getOverlay(): Locator {
    return this.page.getByRole("dialog").filter({
      has: this.getSectionContainer(),
    });
  }
}
