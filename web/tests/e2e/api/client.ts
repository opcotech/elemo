import { TestAuthClient } from "./auth-client";
import { getTestConfig } from "../utils/test-config";

import type { LoginCredentials } from "@/lib/auth/types";
import type { Client } from "@/lib/client/client";
import { createClient } from "@/lib/client/client";

const OAUTH_SCOPES = [
  "user",
  "organization",
  "namespace",
  "project",
  "issue",
  "document",
  "label",
  "todo",
  "notification",
];

let cachedClient: Client | null = null;
let cachedTokens: { access_token: string; refresh_token?: string } | null =
  null;

/**
 * Create an authenticated API client using email and password.
 * Handles OAuth authentication and token management.
 */
export async function createAuthenticatedClient(
  email: string,
  password: string
): Promise<Client> {
  const config = getTestConfig();

  // OAuth endpoint is at root level (/oauth/token), not under /api
  // Extract base URL without /api suffix for OAuth requests
  const oauthBaseUrl = config.apiBaseUrl.replace(/\/api\/?$/, "");

  // Authenticate using OAuth password grant
  const authClient = new TestAuthClient(oauthBaseUrl, {
    clientId: config.authClientId,
    clientSecret: config.authClientSecret,
    tokenUrl: "/oauth/token",
    scopes: OAUTH_SCOPES,
  });

  const credentials: LoginCredentials = { email, password };
  const tokens = await authClient.login(credentials);

  // Create and configure API client
  const client = createClient({
    baseUrl: config.apiBaseUrl,
    // eslint-disable-next-line @typescript-eslint/require-await
    auth: async () => tokens.access_token,
  });

  // Cache tokens for potential refresh
  cachedTokens = {
    access_token: tokens.access_token,
    refresh_token: tokens.refresh_token,
  };

  return client;
}

/**
 * Create an authenticated API client using the e2e privileged user.
 * That user has organization.create on Installation, not a system role.
 */
export async function createPrivilegedClient(): Promise<Client> {
  // Return cached client if available and tokens are still valid
  if (cachedClient && cachedTokens) {
    try {
      const config = getTestConfig();
      const oauthBaseUrl = config.apiBaseUrl.replace(/\/api\/?$/, "");
      const authClient = new TestAuthClient(oauthBaseUrl, {
        clientId: config.authClientId,
        clientSecret: config.authClientSecret,
        tokenUrl: "/oauth/token",
        scopes: OAUTH_SCOPES,
      });
      const isValid = await authClient.validateToken(cachedTokens.access_token);
      if (isValid) {
        return cachedClient;
      }
    } catch {
      // Token validation failed, create new client
    }
  }

  const config = getTestConfig();
  const client = await createAuthenticatedClient(
    config.systemOwnerEmail,
    config.systemOwnerPassword
  );

  cachedClient = client;
  return client;
}

/**
 * Clear cached client and tokens.
 * Useful for testing or when tokens need to be refreshed.
 */
export function clearCachedClient(): void {
  cachedClient = null;
  cachedTokens = null;
}
