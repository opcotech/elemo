import SessionProvider from '@/components/SessionProvider';
import type {AppProps} from 'next/app';
import {Lato, Work_Sans} from 'next/font/google';

import '@/app/globals.css';

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

export default function App({Component, pageProps: {session, ...pageProps}}: AppProps) {
  return (
    <>
      <style jsx global>
        {`
          :root {
            --font-lato: ${lato.style.fontFamily};
            --font-work-sans: ${workSans.style.fontFamily};
          }
        `}
      </style>
      <div className="h-full">
        <SessionProvider session={session}>
          <Component {...pageProps} />
        </SessionProvider>
      </div>
    </>
  );
}
