'use client';

import { setToken } from '@/lib/api';
import { SessionProvider, SessionProviderProps } from 'next-auth/react';

const SESSION_REFETCH_INTERVAL = 60;

export default function Provider(props: SessionProviderProps) {
  setToken(props.session?.user?.access_token || '');

  return (
    <SessionProvider refetchOnWindowFocus={true} refetchInterval={SESSION_REFETCH_INTERVAL} {...props}>
      {props.children}
    </SessionProvider>
  );
}
