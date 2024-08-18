import * as React from 'react';
import NextLink, { LinkProps as NextLinkProps } from 'next/link';
import { cva, type VariantProps } from 'class-variance-authority';

import { cn } from '@/lib/utils';
import { IconExternalLink } from '@tabler/icons-react';

const linkVariants = cva(
  'inline-flex cursor-pointer text-sm transition-colors underline-offset-4 underline decoration-dotted hover:decoration-solid focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:opacity-50',
  {
    variants: {
      variant: {
        default: 'text-primary dark:text-accent',
        destructive: 'text-destructive',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
);

export interface LinkProps extends NextLinkProps, VariantProps<typeof linkVariants> {
  isExternal?: boolean;
  asNextLink?: boolean;
}

const Link = React.forwardRef<HTMLAnchorElement, LinkProps & React.HTMLProps<HTMLAnchorElement>>(
  ({ className, variant, isExternal, asNextLink = false, children, ...props }, ref) => {
    const Comp = asNextLink ? NextLink : 'a';
    return (
      <Comp className={cn(linkVariants({ variant, className }))} ref={ref} {...props}>
        {children}
        {isExternal && <IconExternalLink className='ml-0.5 mt-0.5 h-4 w-4' />}
      </Comp>
    );
  }
);
Link.displayName = 'Button';

export { Link, linkVariants };
