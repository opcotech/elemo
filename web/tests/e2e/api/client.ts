import { getTestConfig } from "../utils/test-config";

import { AuthClient } from "@/lib/auth/auth-client";
import type { LoginCredentials } from "@/lib/auth/types";
import type { Client } from "@/lib/client/client";
import { createClient } from "@/lib/client/client";

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
  const authClient = new AuthClient(oauthBaseUrl, {
    clientId: config.authClientId,
    clientSecret: config.authClientSecret,
    tokenUrl: "/oauth/token",
    scopes: ["user", "organization", "todo", "notification"],
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
 * Create an authenticated API client using system owner credentials.
 * This is the primary method for tests to get an API client.
 */
export async function createSystemOwnerClient(): Promise<Client> {
  // Return cached client if available and tokens are still valid
  if (cachedClient && cachedTokens) {
    try {
      const authClient = new AuthClient(getTestConfig().apiBaseUrl);
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
