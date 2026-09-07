import { test as base } from "@playwright/test";
import type { Client } from "@/lib/api/client";
import {
  createAuthenticatedClient,
  createPrivilegedClient,
} from "../api/client";

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
  // biome-ignore lint/correctness/noEmptyPattern: Playwright fixture with no dependencies
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
  // biome-ignore lint/correctness/noEmptyPattern: Playwright fixture with no dependencies
  createApiClient: async ({}, use) => {
    await use(async (username: string, password: string) => {
      return await createAuthenticatedClient(username, password);
    });
  },
});

export { expect } from "@playwright/test";
