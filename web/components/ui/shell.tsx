import * as React from 'react';

import { cn } from '@/lib/utils';
import { cva, VariantProps } from 'class-variance-authority';

const shellAsideVariants = cva('fixed inset-y-0 top-0 left-0 pt-16 px-1.5 z-10 border-r bg-background', {
  variants: {
    size: {
      default: 'w-18',
      lg: 'w-64',
    },
  },
  defaultVariants: {
    size: 'default',
  },
});

interface ShellAsideProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof shellAsideVariants> {}

const ShellAside = ({ size, className, children, ...props }: ShellAsideProps) => (
  <aside className={cn(shellAsideVariants({ size }), className)} {...props}>
    <div className='flex h-full flex-col items-center gap-4 px-2 py-4'>{children}</div>
  </aside>
);
ShellAside.displayName = 'ShellAside';

const ShellContent = ({ className, children, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <main className={cn('container pt-10', className)} {...props}>
    {children}
  </main>
);
ShellContent.displayName = 'ShellContent';

const Shell = ({ className, children, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <section className={cn('h-full bg-secondary-light dark:bg-muted-dark', className)} {...props}>
    {children}
  </section>
);
Shell.displayName = 'Shell';

export { Shell, ShellAside, ShellContent };
