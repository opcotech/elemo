'use client';

import {SessionProvider, SessionProviderProps} from 'next-auth/react';

const SESSION_REFETCH_INTERVAL = 60;

export default function Provider(props: SessionProviderProps) {
  return (
    <SessionProvider refetchOnWindowFocus={true} refetchInterval={SESSION_REFETCH_INTERVAL} {...props}>
      {props.children}
    </SessionProvider>
  );
}
