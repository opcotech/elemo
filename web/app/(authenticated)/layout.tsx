import { Lato, Work_Sans } from 'next/font/google';
import { MessageArea } from '@/components/MessageArea';
import { Navbar, NavigationItem, UserNavigationItem } from '@/components/navigation/Navbar';
import { NotificationDrawer } from '@/components/notifications';
import { TodoDrawer } from '@/components/todos';
import ErrorBoundary from '@/components/ErrorBoundary';
import Provider from '@/components/Provider';

import '../globals.css';

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
  { id: 'menu-item-logout', label: 'Logout', href: '/api/auth/signout', prefetch: false }
];

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`h-full ${lato.className} ${workSans.className}`}>
      <body className={'h-full'}>
        <ErrorBoundary>
          <Provider>
            <Navbar navigation={navigation} userNavigation={userNavigation} />

            {children}

            <TodoDrawer />
            <NotificationDrawer />
            <MessageArea />
          </Provider>
        </ErrorBoundary>
      </body>
    </html>
  );
}
