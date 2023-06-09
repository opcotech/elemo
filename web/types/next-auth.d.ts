import type { DefaultSession, DefaultUser } from 'next-auth';

interface ElemoUser extends DefaultUser {
  accessToken: string;
  accessTokenExpiresAt: number;
  refreshToken: string;
}

declare module 'next-auth' {
  interface User extends ElemoUser {}

  interface Session extends DefaultSession {
    user?: User;
  }
}

declare module 'next-auth/jwt' {
  interface JWT {
    user?: ElemoUser;
    error?: string;
  }
}
