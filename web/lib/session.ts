import {signOut} from "next-auth/react";
import type {Session} from "next-auth";

export async function getSession(cookie: string): Promise<Session | null> {
  const response = await fetch(`${process.env.NEXTAUTH_URL}/api/auth/session`, {
    headers: {cookie},
  });

  if (!response?.ok) {
    return null;
  }

  const session = await response.json();
  return Object.keys(session).length > 0 ? session : null;
}

export function logout(): void {
  window.localStorage.clear();
  window.sessionStorage.clear();
  signOut();
}