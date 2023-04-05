interface ElemoUser {
}

declare module 'next-auth' {
  interface Account {
    access_token: string;
    refresh_token: string;
    expires_at: number;
  }

  interface Session {
    accessToken?: string;
    user?: AgilexUser;
    error?: string;
  }
}

declare module 'next-auth/jwt/types' {
  interface JWT {
    accessToken?: string;
    accessTokenExpires?: number;
    refreshToken?: string;
    user?: ElemoUser;
    error?: string;
  }
}
