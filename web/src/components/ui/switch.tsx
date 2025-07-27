"use client";

import * as SwitchPrimitive from "@radix-ui/react-switch";
import * as React from "react";

import { cn } from "@/lib/utils";

function Switch({
  className,
  ...props
}: React.ComponentProps<typeof SwitchPrimitive.Root>) {
  return (
    <SwitchPrimitive.Root
      data-slot="switch"
      className={cn(
        "peer data-[state=checked]:bg-primary focus-visible:border-ring focus-visible:ring-ring inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full border transition-all outline-none focus-visible:ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 data-[state=unchecked]:border-gray-300 data-[state=unchecked]:bg-gray-200 dark:data-[state=unchecked]:border-gray-600 dark:data-[state=unchecked]:bg-gray-700",
        className
      )}
      {...props}
    >
      <SwitchPrimitive.Thumb
        data-slot="switch-thumb"
        className={cn(
          "data-[state=checked]:bg-primary-foreground pointer-events-none block size-5 rounded-full bg-white shadow-sm ring-0 transition-transform data-[state=checked]:translate-x-5 data-[state=unchecked]:translate-x-0"
        )}
      />
    </SwitchPrimitive.Root>
  );
}

export { Switch };
