import { cva } from "class-variance-authority";
import type { VariantProps } from "class-variance-authority";
import { Loader2Icon } from "lucide-react";
import * as React from "react";

import { cn } from "@/lib/utils";

const spinnerVariants = cva("animate-spin motion-reduce:animate-none", {
  variants: {
    variant: {
      default: "",
      primary: "text-primary",
      secondary: "text-muted-foreground",
    },
    size: {
      default: "size-6",
      xs: "size-4",
      sm: "size-5",
      lg: "size-8",
      xl: "size-12",
      "2xl": "size-16",
    },
  },
  defaultVariants: {
    variant: "default",
    size: "default",
  },
});

export interface SpinnerProps
  extends
    Omit<React.ComponentProps<"div">, "children">,
    VariantProps<typeof spinnerVariants> {}

function Spinner({ className, variant, size, ...props }: SpinnerProps) {
  return (
    <div
      role="status"
      data-slot="spinner"
      className={cn(spinnerVariants({ variant, size }), className)}
      {...props}
    >
      <Loader2Icon className="h-full w-full" aria-hidden />
      <span className="sr-only">Loading...</span>
    </div>
  );
}

export { Spinner, spinnerVariants };
