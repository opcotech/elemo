import "@tanstack/react-start/server-only";

import { z } from "zod";

const serverEnvSchema = z.object({
  API_BASE_URL: z.url(),
  AUTH_CLIENT_ID: z.string().min(1),
  AUTH_CLIENT_SECRET: z.string().min(1),
});

export interface AuthServerEnv {
  apiBaseUrl: URL;
  clientId: string;
  clientSecret: string;
  /**
   * Whether Set-Cookie should include the Secure attribute.
   * Intentionally NOT derived from NODE_ENV — Vite inlines
   * `process.env.NODE_ENV === "production"` as `true` in production builds,
   * which makes http://localhost E2E break on WebKit/Safari.
   */
  secureCookies: boolean;
}

let cachedEnv: AuthServerEnv | undefined;

function resolveSecureCookies(): boolean {
  const explicit = process.env.ELEMO_COOKIE_SECURE;
  if (explicit === "true") return true;
  if (explicit === "false") return false;
  const appUrl = process.env.APP_URL ?? "";
  return appUrl.startsWith("https://");
}

export function getAuthServerEnv(): AuthServerEnv {
  if (cachedEnv) {
    return cachedEnv;
  }

  const parsed = serverEnvSchema.safeParse(process.env);
  if (!parsed.success) {
    throw new Error(
      `Invalid server authentication environment: ${z.prettifyError(parsed.error)}`
    );
  }

  const apiBaseUrl = new URL(parsed.data.API_BASE_URL);
  if (!["http:", "https:"].includes(apiBaseUrl.protocol)) {
    throw new Error("API_BASE_URL must use http or https");
  }
  if (apiBaseUrl.username || apiBaseUrl.password) {
    throw new Error("API_BASE_URL must not contain credentials");
  }

  cachedEnv = {
    apiBaseUrl,
    clientId: parsed.data.AUTH_CLIENT_ID,
    clientSecret: parsed.data.AUTH_CLIENT_SECRET,
    secureCookies: resolveSecureCookies(),
  };

  return cachedEnv;
}
