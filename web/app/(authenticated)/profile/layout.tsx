import type { ReactNode } from 'react';

export const metadata = {
  title: 'Profile | Elemo'
};

export default async function ProfileLayout({ children }: { children: ReactNode }) {
  return <div className="max-w-6xl mx-auto py-16 lg:flex lg:gap-x-8">{children}</div>;
}
