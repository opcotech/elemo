import type { Metadata } from 'next';
import {
  IconFile,
  IconLayoutKanban,
  IconStack2,
  IconMap,
  IconComponents,
  IconRocket,
  IconSettings,
} from '@tabler/icons-react';

import { cn } from '@/lib/utils';
import { Layout, AsideItem } from '@/components/layouts/with-aside';
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

const layoutAsideItems: AsideItem[] = [
  { icon: <IconStack2 stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Backlog' },
  { icon: <IconLayoutKanban stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Boards' },
  { icon: <IconMap stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Roadmap' },
  { icon: <IconRocket stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Releases' },
  { icon: <IconFile stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Documents' },
  { icon: <IconComponents stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Components' },
  { icon: <IconSettings stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Settings' },
];

export default function RootLayout(props: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang='en'>
      <body className={cn(fontHeading.className, fontBody.className)}>
        <ThemeProvider attribute='class' defaultTheme='system'>
          <Layout {...props} asideItems={layoutAsideItems} />
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}
