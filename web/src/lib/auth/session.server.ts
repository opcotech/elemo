import "@tanstack/react-start/server-only";

import {
  deleteCookie,
  getCookie,
  setCookie,
} from "@tanstack/react-start/server";

import { rotateRefreshToken } from "./oauth.server";
import { getAuthServerEnv } from "./server-env";
import type { AuthTokens } from "./types";

const ACCESS_TOKEN_COOKIE = "elemo_at";
const REFRESH_TOKEN_COOKIE = "elemo_rt";
const DEFAULT_ACCESS_TOKEN_TTL_SECONDS = 60 * 60;
const REFRESH_TOKEN_TTL_SECONDS = 30 * 24 * 60 * 60;

const refreshes = new Map<string, Promise<AuthTokens>>();

function cookieOptions(maxAge: number) {
  return {
    httpOnly: true,
    secure: getAuthServerEnv().secureCookies,
    sameSite: "lax" as const,
    path: "/",
    maxAge,
  };
}

export interface SessionTokens {
  accessToken?: string;
  refreshToken?: string;
}

export function readSessionTokens(): SessionTokens {
  return {
    accessToken: getCookie(ACCESS_TOKEN_COOKIE),
    refreshToken: getCookie(REFRESH_TOKEN_COOKIE),
  };
}

export function writeSessionTokens(
  tokens: AuthTokens,
  previousRefreshToken?: string
): SessionTokens {
  const refreshToken = tokens.refresh_token || previousRefreshToken;

  setCookie(
    ACCESS_TOKEN_COOKIE,
    tokens.access_token,
    cookieOptions(tokens.expires_in || DEFAULT_ACCESS_TOKEN_TTL_SECONDS)
  );

  if (refreshToken) {
    setCookie(
      REFRESH_TOKEN_COOKIE,
      refreshToken,
      cookieOptions(REFRESH_TOKEN_TTL_SECONDS)
    );
  }

  return {
    accessToken: tokens.access_token,
    refreshToken,
  };
}

export function clearSessionCookies(): void {
  const options = {
    path: "/",
    secure: getAuthServerEnv().secureCookies,
    sameSite: "lax" as const,
  };
  deleteCookie(ACCESS_TOKEN_COOKIE, options);
  deleteCookie(REFRESH_TOKEN_COOKIE, options);
}

export async function refreshSession(): Promise<SessionTokens | null> {
  const { refreshToken } = readSessionTokens();
  if (!refreshToken) {
    clearSessionCookies();
    return null;
  }

  let refresh = refreshes.get(refreshToken);
  if (!refresh) {
    refresh = rotateRefreshToken(refreshToken).finally(() => {
      refreshes.delete(refreshToken);
    });
    refreshes.set(refreshToken, refresh);
  }

  try {
    return writeSessionTokens(await refresh, refreshToken);
  } catch {
    clearSessionCookies();
    return null;
  }
}

export async function requireSessionTokens(): Promise<Required<SessionTokens>> {
  const current = readSessionTokens();
  if (current.accessToken && current.refreshToken) {
    return current as Required<SessionTokens>;
  }

  const refreshed = await refreshSession();
  if (!refreshed?.accessToken || !refreshed.refreshToken) {
    throw new AuthenticationError();
  }

  return refreshed as Required<SessionTokens>;
}

export class AuthenticationError extends Error {
  readonly status = 401;

  constructor() {
    super("Authentication required");
    this.name = "AuthenticationError";
  }
}
