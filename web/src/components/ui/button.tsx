import { Slot } from "@radix-ui/react-slot";
import { cva } from "class-variance-authority";
import type { VariantProps } from "class-variance-authority";
import { motion } from "framer-motion";
import * as React from "react";

import { buttonVariants as motionButtonVariants } from "./animations";

import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-sm text-sm font-medium transition-all duration-200 cursor-pointer disabled:pointer-events-none disabled:opacity-50 disabled:cursor-not-allowed [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0 outline-none focus-visible:ring focus-visible:ring-ring focus-visible:ring-offset-2",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground hover:bg-primary/90 focus-visible:ring focus-visible:ring-ring",
        destructive:
          "bg-destructive text-destructive-foreground hover:bg-destructive/90 focus-visible:ring focus-visible:ring-destructive",
        outline:
          "border border-border bg-background text-foreground hover:bg-primary/5 hover:border-primary hover:text-primary focus-visible:ring focus-visible:ring-ring",
        secondary:
          "bg-secondary text-secondary-foreground hover:bg-secondary/80 focus-visible:ring focus-visible:ring-ring",
        ghost:
          "bg-transparent text-foreground hover:bg-primary/5 hover:text-primary focus-visible:ring focus-visible:ring-ring",
        link: "bg-transparent text-primary underline-offset-4 hover:underline focus-visible:ring focus-visible:ring-ring",
        success:
          "bg-success text-success-foreground hover:bg-success/90 focus-visible:ring focus-visible:ring-success",
        warning:
          "bg-warning text-warning-foreground hover:bg-warning/90 focus-visible:ring focus-visible:ring-warning",
      },
      size: {
        default: "h-9 px-4 py-2 has-[>svg]:px-3",
        sm: "text-xs h-8 rounded-sm gap-1.5 px-3 has-[>svg]:px-2.5",
        lg: "h-10 rounded-sm px-6 has-[>svg]:px-4",
        icon: "size-9",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

function Button({
  className,
  variant,
  size,
  asChild = false,
  ...props
}: Omit<
  React.ComponentProps<"button">,
  "onDrag" | "onDragStart" | "onDragEnd"
> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
  }) {
  // If using asChild, we can't animate the child directly, so fallback to normal rendering
  if (asChild) {
    return (
      <Slot
        data-slot="button"
        className={cn(buttonVariants({ variant, size, className }))}
        {...props}
      />
    );
  }

  return (
    <motion.button
      data-slot="button"
      className={cn(buttonVariants({ variant, size, className }))}
      variants={motionButtonVariants}
      initial="initial"
      whileHover="hover"
      whileTap="tap"
      whileFocus="focus"
      {...(props as any)}
    />
  );
}

export { Button, buttonVariants };
