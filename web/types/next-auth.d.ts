import type {DefaultSession, DefaultUser} from 'next-auth';

interface ElemoUser extends DefaultUser {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

declare module 'next-auth' {
  interface User extends ElemoUser {
  }

  interface Session extends DefaultSession {
    accessToken?: string;
  }

  interface Account {
    access_token: string;
    refresh_token: string;
    expires_at: number;
    expires_in: number;
  }
}

declare module 'next-auth/jwt' {
  interface JWT {
    accessToken?: string;
    accessTokenExpires?: number;
    refreshToken?: string;
    user?: ElemoUser;
    error?: string;
  }
}
