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
      await searchInput.fill(term);
    }
  }

  return SearchMixinClass;
}
