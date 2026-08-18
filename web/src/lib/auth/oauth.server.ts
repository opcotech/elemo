import "@tanstack/react-start/server-only";

import { getAuthServerEnv } from "./server-env";
import type { AuthTokens, LoginCredentials } from "./types";

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

export class OAuthError extends Error {
  readonly status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "OAuthError";
    this.status = status;
  }
}

async function requestTokens(body: URLSearchParams): Promise<AuthTokens> {
  const { apiBaseUrl, clientId, clientSecret } = getAuthServerEnv();
  body.set("client_id", clientId);
  body.set("client_secret", clientSecret);

  const response = await fetch(new URL("/oauth/token", apiBaseUrl), {
    method: "POST",
    headers: {
      "content-type": "application/x-www-form-urlencoded",
    },
    body,
    cache: "no-store",
  });

  if (!response.ok) {
    const details = (await response.json().catch(() => null)) as {
      error_description?: unknown;
      message?: unknown;
    } | null;
    const message =
      (typeof details?.error_description === "string" &&
        details.error_description) ||
      (typeof details?.message === "string" && details.message) ||
      "Authentication failed";
    throw new OAuthError(message, response.status);
  }

  const tokens = (await response.json()) as AuthTokens;
  if (!tokens.access_token || !tokens.token_type) {
    throw new OAuthError(
      "OAuth provider returned an invalid token response",
      502
    );
  }

  return tokens;
}

export function loginWithPassword(
  credentials: LoginCredentials
): Promise<AuthTokens> {
  return requestTokens(
    new URLSearchParams({
      grant_type: "password",
      username: credentials.email,
      password: credentials.password,
      scope: OAUTH_SCOPES.join(" "),
    })
  );
}

export function rotateRefreshToken(refreshToken: string): Promise<AuthTokens> {
  return requestTokens(
    new URLSearchParams({
      grant_type: "refresh_token",
      refresh_token: refreshToken,
    })
  );
}
