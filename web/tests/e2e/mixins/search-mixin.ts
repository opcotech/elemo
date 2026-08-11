import { expect } from "@playwright/test";
import type { Locator } from "@playwright/test";

/**
 * Configuration for search mixin.
 */
export interface SearchMixinConfig {
  /**
   * Getter for the section container that contains the search input.
   */
  getSectionContainer: () => Locator;
  /**
   * Placeholder text for the search input.
   */
  searchPlaceholder: string;
}

/**
 * Mixin for sections that have search functionality.
 * Provides common search interaction methods.
 */
export function SearchMixin<T extends abstract new (...args: any[]) => any>(
  Base: T
) {
  abstract class SearchMixinClass extends Base {
    protected searchConfig?: SearchMixinConfig;

    /**
     * Set the search configuration.
     * Should be called in the constructor of the implementing class.
     */
    protected setSearchConfig(config: SearchMixinConfig): void {
      this.searchConfig = config;
    }

    /**
     * Get the search configuration.
     */
    protected getSearchConfig(): SearchMixinConfig {
      if (!this.searchConfig) {
        throw new Error(
          "Search config not set. Call setSearchConfig() in constructor."
        );
      }
      return this.searchConfig;
    }

    /**
     * Get the search input locator.
     */
    getSearchInput(): Locator {
      return this.getSearchConfig()
        .getSectionContainer()
        .getByPlaceholder(this.getSearchConfig().searchPlaceholder);
    }

    /**
     * Fill the search input with a search term.
     */
    async search(term: string): Promise<void> {
      const searchInput = this.getSearchInput();
      await expect(searchInput).toBeVisible();
      await expect(searchInput).toBeEditable();

      // Controlled Base UI inputs on Firefox often ignore fill()/clear();
      // trusted key events keep React filter state in sync.
      for (let attempt = 0; attempt < 2; attempt++) {
        await searchInput.click();
        await searchInput.press("ControlOrMeta+A");
        await searchInput.press("Backspace");
        if (term.length > 0) {
          await searchInput.pressSequentially(term, { delay: 10 });
        }
        try {
          await expect(searchInput).toHaveValue(term, { timeout: 1_000 });
          return;
        } catch {
          if (attempt === 1) {
            throw new Error(
              `Failed to set search input to a stable value after retry`
            );
          }
        }
      }
    }
  }

  return SearchMixinClass;
}
