/**
 * Playwright fixtures for E2E tests.
 * Re-exports test and expect with custom fixtures.
 *
 * Usage in test files:
 * ```typescript
 * import { test, expect } from './fixtures';
 * ```
 */

import { mergeExpects, mergeTests } from "@playwright/test";

import { expect as apiExpect, test as apiTest } from "./api";
import { expect as configExpect, test as configTest } from "./config";

export const test = mergeTests(apiTest, configTest);
export const expect = mergeExpects(apiExpect, configExpect);
