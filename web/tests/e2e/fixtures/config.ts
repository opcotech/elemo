import { test as base } from "@playwright/test";

import { getTestConfig } from "../utils/test-config";
import type { TestConfig } from "../utils/test-config";

/**
 * Custom Playwright fixtures for test configuration.
 * Provides test configuration to all tests.
 */
type ConfigFixtures = {
  testConfig: TestConfig;
};

export const test = base.extend<ConfigFixtures>({
  /**
   * Return the test configuration.
   */
  // eslint-disable-next-line no-empty-pattern
  testConfig: async ({}, use: (config: TestConfig) => Promise<void>) => {
    const config = getTestConfig();
    await use(config);
  },
});

export { expect } from "@playwright/test";
