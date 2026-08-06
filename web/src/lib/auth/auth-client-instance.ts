import { AuthClient } from "./auth-client";
import type { AuthConfig } from "./types";

import { config } from "@/config";

const getAuthConfig = (): AuthConfig => ({
  clientId: config.auth().clientId,
  clientSecret: config.auth().clientSecret,
  tokenUrl: "/oauth/token",
  scopes: ["user", "organization", "todo", "notification"],
});

let instance: AuthClient | null = null;

function getAuthClient(): AuthClient {
  if (!instance) {
    instance = new AuthClient(config.auth().apiBaseUrl, getAuthConfig());
  }
  return instance;
}

/**
 * App-scoped AuthClient singleton. Lazily constructed so importing this module
 * during Node/Playwright analysis does not require Vite env vars.
 */
export const authClient: AuthClient = new Proxy({} as AuthClient, {
  get(_target, prop, receiver) {
    const client = getAuthClient();
    const value = Reflect.get(client, prop, receiver);
    return typeof value === "function" ? value.bind(client) : value;
  },
});
