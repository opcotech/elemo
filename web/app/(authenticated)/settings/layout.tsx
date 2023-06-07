import type { ReactNode } from 'react';
import { Sidebar, type NavigationItem } from '@/components/navigation/Sidebar';

export const metadata = {
  title: 'Settings | Elemo'
};

const sidebarItems: NavigationItem[] = [
  { label: 'General', href: '/settings', icon: 'AdjustmentsHorizontalIcon' },
  { label: 'Security', href: '/settings/security', icon: 'FingerPrintIcon' },
  { label: 'Organizations', href: '/settings/organizations', icon: 'BuildingOffice2Icon' },
  { label: 'System', href: '/settings/system', icon: 'Cog8ToothIcon' }
];

export default async function SettingsLayout({ children }: { children: ReactNode }) {
  return (
    <div className="max-w-6xl mx-auto py-20 lg:flex lg:gap-x-8">
      <Sidebar className="lg:w-1/5 lg:py-2 lg:px-2" navigation={sidebarItems} />
      <main className="lg:w-4/5 py-2 px-2">{children}</main>
    </div>
  );
}
