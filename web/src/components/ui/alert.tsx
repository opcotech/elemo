import { cva } from "class-variance-authority";
import type { VariantProps } from "class-variance-authority";
import * as React from "react";

import { cn } from "@/lib/utils";

const alertVariants = cva(
  "relative w-full rounded-lg border px-4 py-3 text-sm grid has-[>svg]:grid-cols-[calc(var(--spacing)*4)_1fr] grid-cols-[0_1fr] has-[>svg]:gap-x-3 gap-y-0.5 items-start [&>svg]:size-4 [&>svg]:translate-y-0.5 [&>svg]:text-current",
  {
    variants: {
      variant: {
        default: "bg-card text-card-foreground border-border",
        destructive:
          "bg-red-50 text-red-800 border-red-300 dark:bg-red-950/30 dark:text-red-200 dark:border-red-800 shadow-elegant [&>svg]:text-red-600 dark:[&>svg]:text-red-400",
        warning:
          "bg-amber-50 text-amber-900 border-amber-300 dark:bg-amber-900/30 dark:text-amber-200 dark:border-amber-700 shadow-elegant [&>svg]:text-amber-900 dark:[&>svg]:text-amber-200",
        success:
          "bg-green-50 text-green-900 border-green-300 dark:bg-green-900/30 dark:text-green-200 dark:border-green-700 shadow-elegant [&>svg]:text-green-900 dark:[&>svg]:text-green-200",
        info: "bg-blue-50 text-blue-900 border-blue-300 dark:bg-blue-900/30 dark:text-blue-200 dark:border-blue-700 shadow-elegant [&>svg]:text-blue-900 dark:[&>svg]:text-blue-200",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

function Alert({
  className,
  variant,
  ...props
}: React.ComponentProps<"div"> & VariantProps<typeof alertVariants>) {
  return (
    <div
      data-slot="alert"
      role="alert"
      className={cn(alertVariants({ variant }), className)}
      {...props}
    />
  );
}

function AlertTitle({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="alert-title"
      className={cn(
        "col-start-2 line-clamp-1 min-h-4 font-medium tracking-tight",
        className
      )}
      {...props}
    />
  );
}

function AlertDescription({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="alert-description"
      className={cn(
        "col-start-2 grid justify-items-start gap-1 text-sm [&_p]:leading-relaxed",
        className
      )}
      {...props}
    />
  );
}

export { Alert, AlertTitle, AlertDescription };
