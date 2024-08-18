import { Karla, Lexend } from 'next/font/google';

export const fontHeading = Karla({
  variable: '--font-heading',
  weight: ['400', '500', '600', '700'],
  style: ['normal'],
  display: 'swap',
  subsets: ['latin-ext'],
});

export const fontBody = Lexend({
  variable: '--font-body',
  weight: ['300', '400', '500', '600', '700'],
  style: ['normal'],
  display: 'swap',
  subsets: ['latin-ext'],
});
