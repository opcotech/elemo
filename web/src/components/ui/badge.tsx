import { Slot } from "@radix-ui/react-slot";
import { cva } from "class-variance-authority";
import type { VariantProps } from "class-variance-authority";
import * as React from "react";

import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center justify-center rounded-md border px-2.5 py-1 text-xs font-semibold w-fit whitespace-nowrap shrink-0 [&>svg]:size-3 gap-1 [&>svg]:pointer-events-none focus-visible:ring focus-visible:ring-ring focus-visible:ring-offset-2 transition-all duration-200 overflow-hidden cursor-default",
  {
    variants: {
      variant: {
        default:
          "bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/50 dark:text-blue-300 dark:border-blue-800 [a&]:hover:bg-blue-100 [a&]:dark:hover:bg-blue-900/50",
        secondary:
          "bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-900/50 dark:text-gray-300 dark:border-gray-700 [a&]:hover:bg-gray-100 [a&]:dark:hover:bg-gray-800/50",
        destructive:
          "bg-red-50 text-red-700 border-red-200 dark:bg-red-950/50 dark:text-red-300 dark:border-red-800 [a&]:hover:bg-red-100 [a&]:dark:hover:bg-red-900/50",
        outline:
          "bg-background text-foreground border-border [a&]:hover:bg-primary/5 [a&]:hover:text-primary [a&]:hover:border-primary/20",
        success:
          "bg-green-50 text-green-700 border-green-200 dark:bg-green-950/50 dark:text-green-300 dark:border-green-800 [a&]:hover:bg-green-100 [a&]:dark:hover:bg-green-900/50",
        warning:
          "bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/50 dark:text-amber-300 dark:border-amber-800 [a&]:hover:bg-amber-100 [a&]:dark:hover:bg-amber-900/50",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

function Badge({
  className,
  variant,
  asChild = false,
  ...props
}: React.ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & { asChild?: boolean }) {
  const Comp = asChild ? Slot : "span";

  return (
    <Comp
      data-slot="badge"
      className={cn(badgeVariants({ variant }), className)}
      {...props}
    />
  );
}

export { Badge, badgeVariants };
