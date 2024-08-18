import * as React from 'react';
import { IconLoader2 } from '@tabler/icons-react';
import { cva, type VariantProps } from 'class-variance-authority';

import { cn } from '@/lib/utils';

const spinnerVariants = cva('animate-spin motion-reduce:animate-none', {
  variants: {
    variant: {
      default: '',
      primary: 'text-primary dark:text-accent',
    },
    size: {
      default: 'h-6 w-6',
      xs: 'h-4 w-4',
      sm: 'h-5 w-5',
      lg: 'h-8 w-8',
      xl: 'h-12 w-12',
      '2xl': 'h-16 w-16',
    },
  },
  defaultVariants: {
    variant: 'default',
    size: 'default',
  },
});

export interface SpinnerProps
  extends React.HTMLAttributes<Omit<HTMLDivElement, 'children'>>,
    VariantProps<typeof spinnerVariants> {}

function Spinner({ className, variant, size, ...props }: SpinnerProps) {
  return (
    <div role='status' className={cn(spinnerVariants({ variant, size }), className)} {...props}>
      <IconLoader2 className='h-full w-full' />
      <span className='sr-only'>Loading...</span>
    </div>
  );
}

export { Spinner, spinnerVariants };
