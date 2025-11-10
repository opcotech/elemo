import { getTableByHeader, getTableRow } from "../helpers/elements";
import type { Locator, Page } from "@playwright/test";

/**
 * Reusable Table component helper.
 * Provides common table operations.
 */
export class Table {
  constructor(private locator: Locator) {}

  /**
   * Create a Table instance from a page and header text.
   */
  static fromHeader(page: Page, headerText: string): Table {
    return new Table(getTableByHeader(page, headerText));
  }

  /**
   * Get a row by text content.
   */
  getRow(text: string): Locator {
    return getTableRow(this.locator, text);
  }

  /**
   * Get all rows in the table body.
   */
  getRows(): Locator {
    return this.locator.locator("tbody tr");
  }

  /**
   * Get the number of rows in the table.
   */
  async getRowCount(): Promise<number> {
    return await this.getRows().count();
  }

  /**
   * Check if a row exists with the given text.
   */
  async hasRow(text: string): Promise<boolean> {
    const row = this.getRow(text);
    return (await row.count()) > 0;
  }
}
