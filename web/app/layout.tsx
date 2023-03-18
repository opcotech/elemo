import dynamic from 'next/dynamic';
import type { ReactNode } from 'react';
import { Lato, Work_Sans } from 'next/font/google';
import { headers } from 'next/headers';
import ErrorBoundary from '@/components/ErrorBoundary';
import SessionProvider from '@/components/SessionProvider';
import { NavigationItem, UserNavigationItem } from '@/components/Navbar';
import { getSession } from '@/lib/session';

import './globals.css';

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

const navigation: NavigationItem[] = [
  { id: 'menu-item-home', label: 'Home', href: '/', prefetch: true },
  { id: 'menu-item-namespace', label: 'Namespaces', href: '/namespaces', prefetch: true },
  { id: 'menu-item-projects', label: 'Projects', href: '/projects', prefetch: true },
  { id: 'menu-item-documents', label: 'Documents', href: '/documents', prefetch: true }
];

const userNavigation: UserNavigationItem[] = [
  { id: 'menu-item-profile', label: 'Profile', href: '/profile', prefetch: true },
  { id: 'menu-item-settings', label: 'Settings', href: '/settings', prefetch: true },
  { id: 'menu-item-site-settings', label: 'Site settings', href: '/site-settings', prefetch: true },
  { id: 'menu-item-logout', label: 'Logout', href: '/api/auth/signout', prefetch: false }
];

const DynamicNavbar = dynamic(() => import('@/components/Navbar'));
const DynamicTodoDrawer = dynamic(() => import('@/components/Todo'), { ssr: false });
const DynamicNotificationDrawer = dynamic(() => import('@/components/Notification'), { ssr: false });
const DynamicMessageArea = dynamic(() => import('@/components/Message/MessageArea'), { ssr: false });

export default async function RootLayout({ children }: { children: ReactNode }) {
  const session = await getSession(headers().get('cookie') ?? '');

  return (
    <html lang="en" className={`h-full ${lato.className} ${workSans.className}`}>
      <body className={'h-full'}>
        <SessionProvider session={session}>
          <ErrorBoundary>
            <DynamicNavbar navigation={navigation} userNavigation={userNavigation} />

            {children}

            <DynamicTodoDrawer />
            <DynamicNotificationDrawer />
            <DynamicMessageArea />
          </ErrorBoundary>
        </SessionProvider>
      </body>
    </html>
  );
}
