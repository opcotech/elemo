import NextAuth, {DefaultUser} from "next-auth"
import {ISODateString} from "next-auth/src/core/types";

interface ElemoUser extends DefaultUser {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

declare module 'next-auth' {
  interface User extends ElemoUser {
  }

  interface Account {
    access_token: string;
    refresh_token: string;
    expires_at: number;
    expires_in: number;
  }

  interface Session {
    expires: ISODateString
    accessToken?: string;
    user?: ElemoUser;
    error?: string;
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
