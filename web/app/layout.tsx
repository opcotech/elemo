import {headers} from 'next/headers';
import SessionProvider from '@/components/session-provider'
import {getSession} from '@/lib/session';

import './globals.css'

export const metadata = {
  title: 'Elemo',
  description: 'The next-generation project management tool',
}

export default async function RootLayout({children}: { children: React.ReactNode }) {
  const session = await getSession(headers().get('cookie') ?? '');

  return (
    <html lang="en">
    <body>
    <SessionProvider session={session}>
      {children}
    </SessionProvider>
    </body>
    </html>
  )
}
