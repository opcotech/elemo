import * as React from 'react';
import Link, { type LinkProps } from 'next/link';
import { IconMenu } from '@tabler/icons-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';

const NavbarItem = React.forwardRef<
  HTMLAnchorElement,
  React.PropsWithoutRef<LinkProps & React.HTMLProps<HTMLAnchorElement>>
>(({ href, className, children, ...props }, ref) => (
  <Link
    href={href}
    ref={ref}
    className={cn('pt-1 text-sm text-muted-foreground transition-colors hover:text-foreground', className)}
    {...props}
  >
    {children}
  </Link>
));
NavbarItem.displayName = 'NavbarItem';

const NavbarContent = ({ className, children, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <nav
    className={cn(
      'hidden flex-col gap-6 text-lg font-medium md:flex md:flex-row md:items-center md:gap-5 lg:gap-6',
      className
    )}
    {...props}
  >
    <Link href='/' className='text-lg font-semibold md:text-xl'>
      elemo
      <span className='sr-only'>elemo</span>
    </Link>
    {children}
  </nav>
);
NavbarContent.displayName = 'NavbarContent';

const NavbarCollapsedContent = ({ className, children, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <Sheet>
    <SheetTrigger asChild>
      <Button variant='outline' size='icon' className='shrink-0 md:hidden'>
        <IconMenu className='h-5 w-5' />
        <span className='sr-only'>Toggle navigation menu</span>
      </Button>
    </SheetTrigger>
    <SheetContent side='left'>
      <nav className={cn('grid gap-6 text-lg font-medium', className)} {...props}>
        <Link href='/' className='text-lg font-semibold md:text-xl'>
          elemo
          <span className='sr-only'>elemo</span>
        </Link>
        {children}
      </nav>
    </SheetContent>
  </Sheet>
);
NavbarCollapsedContent.displayName = 'NavbarCollapsedContent';

const NavbarActionsContent = ({ className, children, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('ml-auto flex', className)} {...props}>
    <div className='ml-auto flex items-center gap-4 md:gap-2 lg:gap-4'>{children}</div>
  </div>
);
NavbarActionsContent.displayName = 'NavbarActionsContent';

const Navbar = ({ className, children, ...props }: React.HTMLAttributes<HTMLDivElement>) => {
  return (
    <header
      className={cn('sticky top-0 z-50 flex h-16 items-center gap-4 border-b bg-background px-4 md:px-6', className)}
      {...props}
    >
      {children}
    </header>
  );
};
Navbar.displayName = 'Navbar';

export { Navbar, NavbarActionsContent, NavbarCollapsedContent, NavbarContent, NavbarItem };
