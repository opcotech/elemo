import type { AnchorHTMLAttributes } from 'react';

import { concat } from '@/lib/helpers';

export interface LinkProps extends AnchorHTMLAttributes<HTMLAnchorElement> {
  decorated?: boolean;
  prefetch?: boolean;
}

export function Link({ href = '#', prefetch = true, decorated = true, className, children, ...props }: LinkProps) {
  /*
   * FIXME: This component is disabling prefetching for all links, regardless of
   *  the value of the `prefetch` prop. This is because the `prefetch` prop is
   *  storing the resulting page on client-side cache, but the cache is not being
   *  updated when a dynamic route is used. This is a known issue with Next.js:
   *   - https://github.com/vercel/next.js/issues/42991#issuecomment-1517828363
   *  The intended proper workaround would be to use `revalidatePath`, but that
   *  has a bug that prevents client-side cache from being purged if not called
   *  from the same page that is being updated:
   *   - https://github.com/vercel/next.js/issues/42991#issuecomment-1567610032
   * */
  // const Component = href.startsWith('#') ? 'a' : NextLink;
  const Component = 'a';

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
