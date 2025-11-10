import type { Locator } from "@playwright/test";

/**
 * Configuration for table mixin.
 */
export interface TableMixinConfig {
  /**
   * Getter for the section container that contains the table.
   */
  getSectionContainer: () => Locator;
}

/**
 * Mixin for sections that display data in tables.
 * Provides common table interaction methods.
 */
export function TableMixin<T extends abstract new (...args: any[]) => any>(
  Base: T
) {
  abstract class TableMixinClass extends Base {
    protected tableConfig?: TableMixinConfig;

    /**
     * Set the table configuration.
     * Should be called in the constructor of the implementing class.
     */
    protected setTableConfig(config: TableMixinConfig): void {
      this.tableConfig = config;
    }

    /**
     * Get the table configuration.
     */
    protected getTableConfig(): TableMixinConfig {
      if (!this.tableConfig) {
        throw new Error(
          "Table config not set. Call setTableConfig() in constructor."
        );
      }
      return this.tableConfig;
    }

    /**
     * Get the table locator.
     */
    getTable(): Locator {
      return this.getTableConfig().getSectionContainer().getByRole("table");
    }

    /**
     * Get a table row by item name.
     * Can be overridden by subclasses for custom row filtering logic.
     */
    getRowByName(name: string): Locator {
      return this.getTable().getByRole("row").filter({ hasText: name });
    }

    /**
     * Check if a row exists in the table.
     */
    async hasRow(name: string): Promise<boolean> {
      const row = this.getRowByName(name);
      return await row.isVisible().catch(() => false);
    }

    /**
     * Get the count of visible rows in the table.
     */
    async getRowCount(): Promise<number> {
      const table = this.getTable();
      const rows = table.locator("tbody tr");
      return await rows.count();
    }

    /**
     * Get a link within a row by item name.
     */
    getLinkByName(name: string): Locator {
      return this.getRowByName(name).getByRole("link", { name });
    }

    /**
     * Click on a link within a row by item name.
     */
    async clickLink(name: string): Promise<void> {
      const link = this.getLinkByName(name);
      await link.click();
    }
  }

  return TableMixinClass;
}
