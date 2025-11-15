import { test as base } from "@playwright/test";

import {
  createAuthenticatedClient,
  createSystemOwnerClient,
} from "../api/client";

import type { Client } from "@/lib/client/client";

/**
 * Custom Playwright fixtures for API client.
 * Provides authenticated API client to all tests.
 */
type ApiFixtures = {
  systemOwnerApiClient: Client;
  createApiClient: (username: string, password: string) => Promise<Client>;
};

export const test = base.extend<ApiFixtures>({
  /**
   * Authenticated API client using system owner credentials.
   * Automatically created for each test.
   */
  // eslint-disable-next-line no-empty-pattern
  systemOwnerApiClient: async ({}, use: (client: Client) => Promise<void>) => {
    const client = await createSystemOwnerClient();
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
