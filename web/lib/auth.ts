import { NextAuthOptions } from 'next-auth';
import type { JWT } from 'next-auth/jwt';
import Credentials from 'next-auth/providers/credentials';
import { OpenAPI, UsersService } from '@/lib/api';

interface TokenResponse {
  token_type: string;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

interface UserResponse {
  id: string;
  first_name: string | null;
  last_name: string | null;
  email: string;
  picture: string | null;
}

async function getTokenData(credentials: Record<never, string> | undefined): Promise<TokenResponse | null> {
  const payload = {
    ...credentials,
    client_id: process.env.ELEMO_CLIENT_ID || '',
    client_secret: process.env.ELEMO_CLIENT_SECRET || '',
    scope: process.env.ELEMO_AUTH_SCOPES || '',
    grant_type: 'password'
  };

  const tokenResponse = await fetch(`${process.env.NEXT_PUBLIC_ELEMO_BASE_URL}/oauth/token`, {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    method: 'POST',
    body: new URLSearchParams(payload)
  });

  const tokenData: TokenResponse = await tokenResponse.json();

  if (!tokenResponse.ok || !tokenData) {
    return null;
  }

  return tokenData;
}

async function getUserData(tokenData: TokenResponse): Promise<UserResponse | null> {
  OpenAPI.TOKEN = tokenData.access_token;

  const user = await UsersService.v1UserGet('me');

  return {
    id: user.id,
    first_name: user.first_name,
    last_name: user.last_name,
    email: user.email,
    picture: user.picture
  };
}

const ElemoCredentialsProvider = Credentials({
  name: 'Elemo',
  credentials: {},
  authorize: async (credentials) => {
    const tokenData = await getTokenData(credentials);
    if (!tokenData) {
      return null;
    }

    OpenAPI.TOKEN = tokenData.access_token;

    const userData = await getUserData(tokenData);
    if (!userData) {
      return null;
    }

    return {
      id: userData.id,
      name: `${userData.first_name} ${userData.last_name}`,
      email: userData.email,
      image: userData.picture,
      accessToken: tokenData.access_token,
      accessTokenExpiresAt: Date.now() + tokenData.expires_in * 1000,
      refreshToken: tokenData.refresh_token
    };
  }
});

async function refreshAccessToken(token: JWT): Promise<JWT> {
  const response = await fetch(`${process.env.NEXT_PUBLIC_ELEMO_BASE_URL}/oauth/token`, {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    method: 'POST',
    body: new URLSearchParams({
      grant_type: 'refresh_token',
      refresh_token: token.user?.refreshToken ?? '',
      client_id: process.env.ELEMO_CLIENT_ID || '',
      client_secret: process.env.ELEMO_CLIENT_SECRET || '',
      scope: process.env.ELEMO_AUTH_SCOPES || ''
    })
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error_description ?? data.error ?? 'Unknown error');
  }

  OpenAPI.TOKEN = data.access_token;

  return {
    ...{
      ...token,
      user: {
        id: token.user?.id ?? '',
        name: token.user?.name ?? '',
        email: token.user?.email ?? '',
        image: token.user?.image ?? '',
        accessToken: data.access_token,
        accessTokenExpiresAt: Date.now() + data.expires_in * 1000,
        refreshToken: data.refresh_token ?? token.user?.refreshToken
      }
    }
  };
}

export const authOptions: NextAuthOptions = {
  debug: process.env.NODE_ENV === 'development',
  pages: {
    error: '/auth/error',
    signIn: '/auth/signin',
    signOut: '/auth/signout'
  },
  providers: [ElemoCredentialsProvider],
  callbacks: {
    async signIn({ user, account }) {
      if (!user || !account) {
        return false;
      }

      return true;
    },
    async session({ session, token }) {
      return {
        ...session,
        user: token.user,
        error: token.error
      };
    },
    async jwt({ token, user, account }) {
      if (account && user) {
        return { user };
      }

      if (Date.now() < (token.user?.accessTokenExpiresAt ?? 0)) {
        return token;
      }

      try {
        return refreshAccessToken(token);
      } catch (error) {
        console.error(error);
        return { ...token, error: 'RefreshAccessTokenError' };
      }
    }
  }
};
