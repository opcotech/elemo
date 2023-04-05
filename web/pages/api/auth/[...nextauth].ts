import NextAuth from "next-auth/next";
import type {JWT} from "next-auth/jwt/types";
import type {Provider} from "next-auth/providers";
import type {Session} from "next-auth/core/types";

export const ElemoProvider: Provider = {
  id: "elemo",
  name: "Elemo",
  type: "oauth",
  client: {
    client_id: process.env.ELEMO_CLIENT_ID,
    client_secret: process.env.ELEMO_CLIENT_SECRET,
    token_endpoint_auth_method: "client_secret_post"
  },
  authorization: {
    url: `${process.env.ELEMO_BASE_URL}/oauth/authorize`,
    params: {
      scope: process.env.ELEMO_AUTH_SCOPES,
    },
  },
  token: `${process.env.ELEMO_BASE_URL}/oauth/token`,
  userinfo: `${process.env.ELEMO_BASE_URL}/v1/users/me`,
  checks: ["pkce", "state", "nonce"],
  profile: (profile) => {
    console.log('returning', {
      id: profile.id,
      name: `${profile.first_name} ${profile.last_name}`,
      email: profile.email,
      image: profile.picture,
    });

    return {
      id: profile.id,
      name: `${profile.first_name} ${profile.last_name}`,
      email: profile.email,
      image: profile.picture,
    };
  },
};

async function refreshAccessToken(token: JWT): Promise<JWT> {
  const response = await fetch(`${process.env.ELEMO_BASE_URL}/oauth/token`, {
    headers: {"Content-Type": "application/x-www-form-urlencoded"},
    method: "POST",
    body: new URLSearchParams({
      grant_type: "refresh_token",
      refresh_token: token.refreshToken ?? "",
      client_id: process.env.ELEMO_CLIENT_ID || "",
      client_secret: process.env.ELEMO_CLIENT_SECRET || "",
      scope: process.env.ELEMO_AUTH_SCOPES || "",
    }),
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error_description ?? data.error ?? "Unknown error");
  }

  return {
    ...token,
    accessToken: data.access_token,
    accessTokenExpires: Date.now() + data.expires_in * 1000,
    refreshToken: data.refresh_token ?? token.refreshToken,
  };
}

export default NextAuth({
  providers: [
    ElemoProvider,
  ],
  callbacks: {
    async session({session, token}): Promise<Session> {
      return {
        ...session,
        accessToken: token.accessToken,
        user: token.user,
        error: token.error,
      };
    },
    async jwt({token, user, account}): Promise<JWT> {
      if (account && user) {
        return {
          accessToken: account.access_token,
          accessTokenExpires: account.expires_at * 1000,
          refreshToken: account.refresh_token,
          user,
        };
      }

      if (token.accessTokenExpires && Date.now() < token.accessTokenExpires) {
        return token;
      }

      try {
        return refreshAccessToken(token);
      } catch (error) {
        console.error(error);
        return {...token, error: "RefreshAccessTokenError"};
      }
    },
  },
  debug: process.env.NODE_ENV !== "production" && process.env.DEBUG === "true",
});
