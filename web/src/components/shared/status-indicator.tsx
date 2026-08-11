import { cva } from "class-variance-authority";
import type { VariantProps } from "class-variance-authority";
import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

const statusIndicatorVariants = cva(
  "inline-flex w-fit shrink-0 items-center gap-1.5 text-xs font-medium whitespace-nowrap capitalize",
  {
    variants: {
      tone: {
        neutral: "text-muted-foreground",
        primary: "text-primary-on-subtle",
        success: "text-success",
        warning: "text-warning-on-subtle",
        danger: "text-destructive",
        info: "text-info",
      },
    },
    defaultVariants: {
      tone: "neutral",
    },
  }
);

export type StatusIndicatorTone = NonNullable<
  VariantProps<typeof statusIndicatorVariants>["tone"]
>;

function resolveStatusTone(status: string): StatusIndicatorTone {
  const normalized = status.toLowerCase();

  if (
    normalized === "blocked" ||
    normalized === "canceled" ||
    normalized === "cancelled" ||
    normalized === "failed" ||
    normalized === "critical"
  ) {
    return "danger";
  }

  if (
    normalized === "done" ||
    normalized === "active" ||
    normalized === "completed" ||
    normalized === "success"
  ) {
    return "success";
  }

  if (
    normalized === "in-progress" ||
    normalized === "review" ||
    normalized === "high" ||
    normalized === "urgent"
  ) {
    return "primary";
  }

  if (
    normalized === "planned" ||
    normalized === "warning" ||
    normalized === "medium" ||
    normalized === "pending"
  ) {
    return "warning";
  }

  if (normalized === "info" || normalized === "open") {
    return "info";
  }

  return "neutral";
}

export function StatusIndicator({
  status,
  tone,
  className,
  ...props
}: Omit<ComponentProps<"span">, "children"> & {
  status: string;
  tone?: StatusIndicatorTone;
}) {
  const resolvedTone = tone ?? resolveStatusTone(status);

  return (
    <span
      data-slot="status-indicator"
      data-tone={resolvedTone}
      className={cn(statusIndicatorVariants({ tone: resolvedTone }), className)}
      {...props}
    >
      <span
        aria-hidden="true"
        className="size-2 rounded-full bg-current"
        data-slot="status-indicator-dot"
      />
      {status.replaceAll("-", " ")}
    </span>
  );
}

export { statusIndicatorVariants };
