import {headers} from 'next/headers';
import SessionProvider from '@/components/SessionProvider';
import {getSession} from '@/lib/session';
import type {ReactNode} from 'react';

import './globals.css';
import {Lato, Work_Sans} from 'next/font/google';

export const metadata = {
  title: 'Elemo',
  description: 'The next-generation project management tool'
};

const lato = Lato({
  variable: '--font-lato',
  weight: ['300', '400', '700'],
  style: ['normal'],
  display: 'swap',
  subsets: ['latin-ext']
});

const workSans = Work_Sans({
  variable: '--font-work-sans',
  weight: ['300', '400', '500', '700'],
  style: ['normal'],
  display: 'swap',
  subsets: ['latin-ext']
});

export default async function RootLayout({ children }: { children: ReactNode }) {
  const session = await getSession(headers().get('cookie') ?? '');

  return (
    <html lang="en" className={`h-full ${lato.className} ${workSans.className}`}>
      <body className={'h-full'}>
        <SessionProvider session={session}>{children}</SessionProvider>
      </body>
    </html>
  );
}
