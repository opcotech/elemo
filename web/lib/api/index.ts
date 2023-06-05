import { ApiError, OpenAPI } from 'elemo-client';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getSession } from 'next-auth/react';

OpenAPI.BASE = process.env.NEXT_PUBLIC_ELEMO_BASE_URL ?? '';

// Set the token for the API client using the backend session if it exists,
// otherwise use the frontend session assuming we are on the client side.
OpenAPI.TOKEN = async (): Promise<string> => {
  let session = null;

  try {
    session = await getServerSession(authOptions);
  } catch (e) {
    session = await getSession();
  }

  return session?.accessToken || '';
};

export function getErrorMessage(e: unknown) {
  return (e as ApiError).message;
}

export * from 'elemo-client';
