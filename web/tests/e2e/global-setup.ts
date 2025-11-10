import type { FullConfig } from "@playwright/test";

import { verifyBackendAPI } from "./utils/api";
import { getTestConfig } from "./utils/test-config";

/**
 * Playwright global setup hook.
 * Ensures system owner user exists before all tests run.
 * This is the only acceptable use of direct database writes - it's infrastructure, not test data.
 */
// eslint-disable-next-line @typescript-eslint/no-unused-vars
async function globalSetup(_: FullConfig) {
  const testConfig = getTestConfig();

  await verifyBackendAPI(testConfig);
}

export default globalSetup;
