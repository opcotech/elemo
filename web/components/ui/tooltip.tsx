'use client';

import * as React from 'react';
import * as TooltipPrimitive from '@radix-ui/react-tooltip';
import { cva, VariantProps } from 'class-variance-authority';

import { cn } from '@/lib/utils';

const TooltipProvider = TooltipPrimitive.Provider;
const Tooltip = TooltipPrimitive.Root;
const TooltipTrigger = TooltipPrimitive.Trigger;

const tooltipVariants = cva(
  'z-50 overflow-hidden rounded-md px-3 py-2 text-xs animate-in fade-in-0 zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
  {
    variants: {
      variant: {
        default:
          'bg-[#f9fafb] dark:bg-[#9ca3af]/10 text-[#4b5563] dark:text-[#9ca3af] border border-[#6b7280]/10 dark:border-[#9ca3af]/10',
        red: 'bg-[#fef2f2] dark:bg-[#f87171]/10 text-[#b91c1c] dark:text-[#f87171] border border-[#dc2626]/10 dark:border-[#f87171]/10',
        yellow:
          'bg-[#fefce8] dark:bg-[#facc15]/10 text-[#854d0e] dark:text-[#eab308] border border-[#ca8a04]/10 dark:border-[#facc15]/20',
        green:
          'bg-[#f0fdf4] dark:bg-[#22c55e]/10 text-[#15803d] dark:text-[#4ade80] border border-[#16a34a]/10 dark:border-[#22c55e]/20',
        blue: 'bg-[#f0f9ff] dark:bg-[#38bdf8]/10 text-[#0369a1] dark:text-[#38bdf8] border border-[#0369a1]/10 dark:border-[#38bdf8]/30',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
);

interface TooltipContentProps
  extends React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Content>,
    VariantProps<typeof tooltipVariants> {}

const TooltipContent = React.forwardRef<React.ElementRef<typeof TooltipPrimitive.Content>, TooltipContentProps>(
  ({ className, children, variant, sideOffset = 4, ...props }, ref) => (
    <TooltipPrimitive.Content ref={ref} sideOffset={sideOffset} asChild {...props}>
      <div className='z-50 rounded-md bg-background'>
        <div className={cn(tooltipVariants({ variant }), className)}>{children}</div>
      </div>
    </TooltipPrimitive.Content>
  )
);
TooltipContent.displayName = TooltipPrimitive.Content.displayName;

export { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider };
