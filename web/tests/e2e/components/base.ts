import type { Locator, Page } from "@playwright/test";

/**
 * Base component class that provides common functionality.
 * All components should extend this or use composition with it.
 */
export abstract class BaseComponent {
  constructor(protected page: Page) {}

  /**
   * Get the page instance.
   */
  getPage(): Page {
    return this.page;
  }
}

/**
 * Base class for components that wrap a specific locator.
 * Useful for components that represent a specific DOM element.
 */
export abstract class LocatorComponent extends BaseComponent {
  constructor(
    page: Page,
    protected locator: Locator
  ) {
    super(page);
  }

  /**
   * Get the root locator for this component.
   */
  getLocator(): Locator {
    return this.locator;
  }

  /**
   * Check if the component is visible.
   */
  async isVisible(): Promise<boolean> {
    return await this.locator.isVisible().catch(() => false);
  }

  /**
   * Wait for the component to be visible.
   */
  async waitForVisible(options?: { timeout?: number }): Promise<void> {
    await this.locator.waitFor({
      state: "visible",
      timeout: options?.timeout ?? 5000,
    });
  }
}
