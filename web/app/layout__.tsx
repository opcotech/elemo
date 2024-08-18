import type { Metadata } from 'next';

import { cn } from '@/lib/utils';
import { Layout } from '@/components/layouts/default';
import { ThemeProvider } from '@/components/providers/theme-provider';
import { Toaster } from '@/components/ui/toaster';
import { fontHeading, fontBody } from '@/config/fonts';
import { siteConfig } from '@/config/site';

import './styles/globals.css';

export const metadata: Metadata = {
  metadataBase: new URL(siteConfig.frontendUrl),
  title: {
    default: `Home | ${siteConfig.name}`,
    template: `%s | ${siteConfig.name}`,
  },
  description: siteConfig.description,
  icons: {
    icon: '/favicon.png',
    shortcut: '/favicon-16x16.png',
    apple: '/apple-touch-icon.png',
  },
};

export default function RootLayout(props: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang='en'>
      <body className={cn(fontHeading.className, fontBody.className)}>
        <ThemeProvider attribute='class' defaultTheme='system'>
          <Layout {...props} />
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}
