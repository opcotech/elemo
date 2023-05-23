'use client';

import Link from '@/components/Link';
import Page from '@/components/Page';
import { signIn, signOut, useSession } from 'next-auth/react';

export default function Home() {
  const { data: session } = useSession();

  return (
    <Page title='Dashboard'>
      <Link href="/organizations">Organizations</Link>
    </Page>
  );
}
