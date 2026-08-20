import { test as base } from "@playwright/test";

import {
  createAuthenticatedClient,
  createPrivilegedClient,
} from "../api/client";

import type { Client } from "@/lib/client/client";

/**
 * Custom Playwright fixtures for API client.
 * Provides authenticated API client to all tests.
 */
type ApiFixtures = {
  privilegedApiClient: Client;
  createApiClient: (username: string, password: string) => Promise<Client>;
};

export const test = base.extend<ApiFixtures>({
  /**
   * Authenticated API client using the e2e privileged user
   * (organization.create on Installation).
   */
  // eslint-disable-next-line no-empty-pattern
  privilegedApiClient: async ({}, use: (client: Client) => Promise<void>) => {
    const client = await createPrivilegedClient();
    await use(client);
  },

  /**
   * Create an authenticated API client with custom credentials.
   *
   * @param username - User username
   * @param password - User password
   * @returns API client
   */
  // eslint-disable-next-line no-empty-pattern
  createApiClient: async ({}, use) => {
    await use(async (username: string, password: string) => {
      return await createAuthenticatedClient(username, password);
    });
  },
});

export { expect } from "@playwright/test";
