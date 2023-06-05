import { Lato, Work_Sans } from 'next/font/google';
import ErrorBoundary from '@/components/ErrorBoundary';
import Provider from '@/components/Provider';

import '../globals.css';

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

export const metadata = {
  title: 'Elemo',
  description: 'The next-generation project management tool'
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`h-full ${lato.className} ${workSans.className}`}>
      <body className={'h-full'}>
        <ErrorBoundary>
          <Provider>{children}</Provider>
        </ErrorBoundary>
      </body>
    </html>
  );
}
