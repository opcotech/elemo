import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";

import { loginWithPassword } from "./oauth.server";
import { getAuthServerEnv } from "./server-env";
import {
  clearSessionCookies,
  readSessionTokens,
  refreshSession,
  writeSessionTokens,
} from "./session.server";
import type { User } from "./types";

export interface SessionView {
  user: User;
}

const credentialsSchema = z.object({
  email: z.email(),
  password: z.string().min(1).max(256),
});

async function fetchCurrentUser(accessToken: string): Promise<Response> {
  const { apiBaseUrl } = getAuthServerEnv();
  return fetch(new URL("/v1/users/me", apiBaseUrl), {
    headers: {
      authorization: `Bearer ${accessToken}`,
      accept: "application/json",
    },
    cache: "no-store",
  });
}

async function resolveCurrentSession(
  accessTokenOverride?: string
): Promise<SessionView | null> {
  // Prefer an explicit token from the current request (e.g. just after login).
  // setCookie writes response headers; getCookie only sees the incoming request,
  // so freshly written cookies are not readable in the same handler.
  let accessToken = accessTokenOverride ?? readSessionTokens().accessToken;
  if (!accessToken) {
    accessToken = (await refreshSession())?.accessToken;
  }
  if (!accessToken) {
    return null;
  }

  let response = await fetchCurrentUser(accessToken);
  if (response.status === 401) {
    const refreshed = await refreshSession();
    if (!refreshed?.accessToken) {
      return null;
    }
    response = await fetchCurrentUser(refreshed.accessToken);
  }

  if (!response.ok) {
    if (response.status === 401) {
      clearSessionCookies();
      return null;
    }
    throw new Error("Unable to load the current user");
  }

  return {
    user: (await response.json()) as User,
  };
}

export const loginFn = createServerFn({ method: "POST" })
  .validator(credentialsSchema)
  .handler(async ({ data }) => {
    const tokens = await loginWithPassword(data);
    const sessionTokens = writeSessionTokens(tokens);

    try {
      const session = await resolveCurrentSession(sessionTokens.accessToken);
      if (!session) {
        throw new Error("Unable to establish an authenticated session");
      }
      return session;
    } catch (error) {
      clearSessionCookies();
      throw error;
    }
  });

export const currentSessionFn = createServerFn({ method: "GET" }).handler(() =>
  resolveCurrentSession()
);

export const refreshSessionFn = createServerFn({ method: "POST" }).handler(
  async () => {
    const refreshed = await refreshSession();
    if (!refreshed?.accessToken) {
      return null;
    }
    return resolveCurrentSession(refreshed.accessToken);
  }
);

export const logoutFn = createServerFn({ method: "POST" }).handler(() => {
  clearSessionCookies();
  return { authenticated: false as const };
});
