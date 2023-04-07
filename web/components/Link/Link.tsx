import { default as NextLink } from 'next/link';
import type { AnchorHTMLAttributes } from 'react';

import { concat } from '@/helpers';

export interface LinkProps extends AnchorHTMLAttributes<HTMLAnchorElement> {
  decorated?: boolean;
  prefetch?: boolean;
}

export default function Link({
  href = '#',
  prefetch = true,
  decorated = true,
  className,
  children,
  ...props
}: LinkProps) {
  const Component = href.startsWith('#') ? 'a' : NextLink;
  return (
    <Component
      href={href}
      className={concat(className, decorated ? 'link decorated' : 'link')}
      {...props}
      {...(!href.startsWith('#') && !prefetch && { prefetch: prefetch })}
    >
      {children}
    </Component>
  );
}
