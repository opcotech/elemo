'use client';

import { usePathname } from 'next/navigation';
import { toCapitalCase } from '@/lib/helpers';
import { Icon } from '@/components/blocks/Icon';
import { Link } from '@/components/blocks/Link';

export interface BreadcrumbLink {
  name: string;
  href: string;
  current?: boolean;
}

export interface BreadcrumbsProps {
  links?: BreadcrumbLink[];
}

export function Breadcrumbs({ links = [] }: BreadcrumbsProps) {
  const pathname = usePathname();

  if (links.length > 0) {
    links.map((link) => {
      if (link.href === pathname || link.current) {
        link.current = true;
      }
    });
  } else {
    const pathSplit = pathname?.split('/') || [];
    for (let i = 1; i < pathSplit?.length; i++) {
      const path = pathSplit.slice(0, i + 1).join('/');
      links.push({
        name: toCapitalCase(pathSplit[i].replace('-', ' ')),
        href: path,
        current: i === pathSplit.length - 1
      });
    }
  }

  return (
    <nav className="flex" aria-label="Breadcrumb">
      <ol role="list" className="flex items-center space-x-2">
        <li>
          <div>
            <Link href="/" className="text-gray-400 hover:text-gray-500">
              <Icon variant={'HomeIcon'} size="sm" className={'flex-shrink-0 text-gray-600'} aria-hidden={true} />
              <span className="sr-only">Home</span>
            </Link>
          </div>
        </li>
        {links.map((link) => (
          <li key={link.name}>
            <div className="flex items-center">
              <svg
                className="h-4 w-4 flex-shrink-0 text-gray-400"
                fill="currentColor"
                viewBox="0 0 20 20"
                aria-hidden="true"
              >
                <path d="M5.555 17.776l8-16 .894.448-8 16-.894-.448z" />
              </svg>
              {link.current ? (
                <span className="ml-2 text-sm font-medium text-gray-600">{link.name}</span>
              ) : (
                <Link
                  href={link.href}
                  className="ml-2 text-sm font-medium text-gray-600 hover:text-gray-700"
                  aria-current={link.current ? 'page' : undefined}
                >
                  {link.name}
                </Link>
              )}
            </div>
          </li>
        ))}
      </ol>
    </nav>
  );
}
