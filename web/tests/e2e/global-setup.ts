import type { FullConfig } from "@playwright/test";

import { verifyBackendAPI } from "./utils/api";
import { ensurePrivilegedUser } from "./utils/db";
import { getTestConfig } from "./utils/test-config";

/**
 * Playwright global setup hook.
 * Ensures the e2e privileged user exists with organization.create on Installation.
 * This is the only acceptable use of direct database writes - it's infrastructure, not test data.
 */
// eslint-disable-next-line @typescript-eslint/no-unused-vars
async function globalSetup(_: FullConfig) {
  const testConfig = getTestConfig();

  await verifyBackendAPI(testConfig);
  await ensurePrivilegedUser(testConfig);
}

export default globalSetup;
