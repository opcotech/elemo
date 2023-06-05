import { default as NextLink } from 'next/link';
import type { AnchorHTMLAttributes } from 'react';

import { concat } from '@/lib/helpers';

export interface LinkProps extends AnchorHTMLAttributes<HTMLAnchorElement> {
  decorated?: boolean;
  prefetch?: boolean;
}

export function Link({ href = '#', prefetch = true, decorated = true, className, children, ...props }: LinkProps) {
  const Component = href.startsWith('#') ? 'a' : NextLink;

  return (
    <Component
      href={href}
      className={concat(
        'cursor-pointer underline-offset-4 hover:text-blue-400',
        decorated ? 'underline decoration-dashed hover:decoration-solid' : '',
        className
      )}
      {...props}
      {...(!href.startsWith('#') && !prefetch && { prefetch: prefetch })}
    >
      {children}
    </Component>
  );
}
