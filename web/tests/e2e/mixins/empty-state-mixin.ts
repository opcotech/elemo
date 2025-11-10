import type { Locator } from "@playwright/test";

import { waitForElementVisible } from "../helpers";

/**
 * Configuration for empty state mixin.
 */
export interface EmptyStateMixinConfig {
  /**
   * Getter for the section container.
   */
  getSectionContainer: () => Locator;
  /**
   * Text that appears in the empty state.
   */
  emptyStateText: string;
  /**
   * Optional: Getter for the table locator (if applicable).
   * Used to wait for either table or empty state.
   */
  getTable?: () => Locator;
}

/**
 * Mixin for sections that display empty states.
 * Provides common empty state interaction methods.
 */
export function EmptyStateMixin<T extends abstract new (...args: any[]) => any>(
  Base: T
) {
  abstract class EmptyStateMixinClass extends Base {
    protected emptyStateConfig?: EmptyStateMixinConfig;

    /**
     * Set the empty state configuration.
     * Should be called in the constructor of the implementing class.
     */
    protected setEmptyStateConfig(config: EmptyStateMixinConfig): void {
      this.emptyStateConfig = config;
    }

    /**
     * Get the empty state configuration.
     */
    protected getEmptyStateConfig(): EmptyStateMixinConfig {
      if (!this.emptyStateConfig) {
        throw new Error(
          "Empty state config not set. Call setEmptyStateConfig() in constructor."
        );
      }
      return this.emptyStateConfig;
    }

    /**
     * Get the empty state locator.
     */
    getEmptyState(): Locator {
      return (this as any).page.getByText(
        this.getEmptyStateConfig().emptyStateText
      );
    }

    /**
     * Check if empty state is visible.
     */
    async hasEmptyState(): Promise<boolean> {
      try {
        const emptyState = this.getEmptyState();
        return await emptyState.isVisible({ timeout: 2000 });
      } catch {
        return false;
      }
    }

    /**
     * Wait for either the table or empty state to be visible.
     * Useful in waitForLoad methods.
     */
    async waitForTableOrEmptyState(options?: {
      timeout?: number;
    }): Promise<void> {
      const config = this.getEmptyStateConfig();
      const table = config.getTable?.();
      const emptyState = this.getEmptyState();

      if (table) {
        await waitForElementVisible(table.or(emptyState), options);
      } else {
        await waitForElementVisible(emptyState, options);
      }
    }
  }

  return EmptyStateMixinClass;
}
