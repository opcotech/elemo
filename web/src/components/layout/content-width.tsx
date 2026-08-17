import { cva } from "class-variance-authority";
import type { VariantProps } from "class-variance-authority";
import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

const contentWidthVariants = cva(
  "mx-auto w-full px-4 py-6 sm:px-6 lg:px-8 lg:py-8",
  {
    variants: {
      width: {
        document: "max-w-5xl",
        entity: "max-w-5xl",
        overview: "max-w-7xl",
        settings: "max-w-6xl",
        projection: "max-w-none",
        board: "max-w-none",
        table: "max-w-none",
        timeline: "max-w-none",
        graph: "max-w-none",
        full: "max-w-none p-0 sm:p-0 lg:p-0",
      },
    },
    defaultVariants: {
      width: "overview",
    },
  }
);

type ContentWidthProps = ComponentProps<"div"> &
  VariantProps<typeof contentWidthVariants>;

function ContentWidth({ className, width, ...props }: ContentWidthProps) {
  return (
    <div
      data-slot="content-width"
      data-content-width={width ?? "overview"}
      className={cn(contentWidthVariants({ width }), className)}
      {...props}
    />
  );
}

export { ContentWidth, contentWidthVariants };
export type { ContentWidthProps };
