import type { Locator } from "@playwright/test";

import { waitForElementVisible } from "../helpers";

/**
 * Mixin for sections that have a container locator.
 * Provides common functionality for waiting for section containers to load.
 */
export function SectionContainerMixin<
  T extends abstract new (...args: any[]) => any,
>(Base: T) {
  abstract class SectionContainerMixinClass extends Base {
    protected sectionContainer?: Locator;

    /**
     * Set the section container locator.
     * Should be called in the constructor of the implementing class.
     */
    protected setSectionContainer(container: Locator): void {
      this.sectionContainer = container;
    }

    /**
     * Get the section container locator.
     */
    getSectionContainer(): Locator {
      if (!this.sectionContainer) {
        throw new Error(
          "Section container not set. Call setSectionContainer() in constructor."
        );
      }
      return this.sectionContainer;
    }

    /**
     * Wait for section container to load and be visible.
     */
    async waitForContainerLoad(options?: { timeout?: number }): Promise<void> {
      await waitForElementVisible(this.getSectionContainer(), options);
    }
  }

  return SectionContainerMixinClass;
}
