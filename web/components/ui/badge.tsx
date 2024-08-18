import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';

import { cn } from '@/lib/utils';

const badgeVariants = cva(
  'inline-flex items-center lowercase rounded px-2.5 py-0.5 text-xs font-medium focus:outline-none ring-1 ring-inset',
  {
    variants: {
      variant: {
        default:
          'bg-[#f9fafb] dark:bg-[#9ca3af]/10 text-[#4b5563] dark:text-[#9ca3af] ring-[#6b7280]/10 dark:ring-[#9ca3af]/10',
        red: 'bg-[#fef2f2] dark:bg-[#f87171]/10 text-[#b91c1c] dark:text-[#f87171] ring-[#dc2626]/10 dark:ring-[#f87171]/10',
        yellow:
          'bg-[#fefce8] dark:bg-[#facc15]/10 text-[#854d0e] dark:text-[#eab308] ring-[#ca8a04]/10 dark:ring-[#facc15]/20',
        green:
          'bg-[#f0fdf4] dark:bg-[#22c55e]/10 text-[#15803d] dark:text-[#4ade80] ring-[#16a34a]/10 dark:ring-[#22c55e]/20',
        blue: 'bg-[#f0f9ff] dark:bg-[#38bdf8]/10 text-[#0369a1] dark:text-[#38bdf8] ring-[#0369a1]/10 dark:ring-[#38bdf8]/30',
        purple:
          'bg-[#faf5ff] dark:bg-[#c084fc]/10 text-[#7e22ce] dark:text-[#c084fc] ring-[#7e22ce]/10 dark:ring-[#c084fc]/30',
        pink: 'bg-[#fdf2f8] dark:bg-[#f472b6]/10 text-[#be185d] dark:text-[#f472b6] ring-[#be185d]/10 dark:ring-[#f472b6]/30',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
);

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />;
}
Badge.displayName = 'Badge';

export { Badge, badgeVariants };
