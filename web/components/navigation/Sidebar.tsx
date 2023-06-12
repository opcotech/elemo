'use client';

import { HTMLAttributes, useCallback } from 'react';
import { usePathname } from 'next/navigation';
import type { IconVariant } from '@/types/heroicon';
import { concat } from '@/lib/helpers';
import { Icon } from '@/components/blocks/Icon';
import { Link } from '@/components/blocks/Link';

export interface NavigationItem {
  label: string;
  href: string;
  icon?: IconVariant;
}

export interface SidebarProps extends HTMLAttributes<HTMLElement> {
  navigation: NavigationItem[];
}

export function Sidebar({ navigation, ...props }: SidebarProps) {
  const currentPath = usePathname();

  const isCurrent = useCallback(
    (href: string) => {
      return currentPath === href;
    },
    [currentPath]
  );

  const generateId = useCallback((href: string) => {
    return href.replace('/', '-').toLowerCase();
  }, []);

  return (
    <aside className="flex overflow-x-auto border-b border-gray-900/5 lg:block lg:flex-none lg:border-0" {...props}>
      <nav className="flex-none px-4 sm:px-6 lg:px-0">
        <ul role="list" className="flex gap-x-3 gap-y-1 whitespace-nowrap lg:flex-col">
          {navigation.map((item) => (
            <li key={item.label}>
              <Link
                id={generateId(item.href)}
                key={item.label}
                href={item.href}
                decorated={false}
                prefetch={!isCurrent(item.href)}
                className={concat(
                  isCurrent(item.href)
                    ? 'bg-gray-50 text-blue-500'
                    : 'text-gray-700 hover:text-blue-500 hover:bg-gray-50',
                  'group flex gap-x-3 rounded-md py-3 pl-2 pr-3 text-sm font-medium'
                )}
              >
                {item.icon && (
                  <Icon
                    size="xs"
                    variant={item.icon}
                    className={concat(
                      isCurrent(item.href) ? 'text-blue-500' : 'text-gray-400 group-hover:text-blue-500',
                      'h-5 w-5'
                    )}
                  />
                )}

                {item.label}
              </Link>
            </li>
          ))}
        </ul>
      </nav>
    </aside>
  );
}
