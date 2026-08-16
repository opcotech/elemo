import { cva } from "class-variance-authority";
import type { VariantProps } from "class-variance-authority";
import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

const statusIndicatorVariants = cva("inline-flex items-center gap-2", {
  variants: {
    tone: {
      neutral: "text-foreground",
      primary: "text-primary",
      success: "text-success",
      warning: "text-warning-on-subtle",
      danger: "text-destructive",
      info: "text-info",
    },
  },
  defaultVariants: {
    tone: "neutral",
  },
});

export type StatusIndicatorTone = NonNullable<
  VariantProps<typeof statusIndicatorVariants>["tone"]
>;

export function resolveStatusTone(status: string): StatusIndicatorTone {
  const normalized = status.toLowerCase();

  if (
    normalized === "blocked" ||
    normalized === "canceled" ||
    normalized === "cancelled" ||
    normalized === "closed" ||
    normalized === "failed" ||
    normalized === "critical" ||
    normalized === "highest"
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
    normalized === "in progress" ||
    normalized === "high" ||
    normalized === "urgent"
  ) {
    return "primary";
  }

  if (
    normalized === "planned" ||
    normalized === "in review" ||
    normalized === "warning" ||
    normalized === "medium" ||
    normalized === "normal" ||
    normalized === "review" ||
    normalized === "pending"
  ) {
    return "warning";
  }

  if (
    normalized === "info" ||
    normalized === "open" ||
    normalized === "backlog"
  ) {
    return "info";
  }

  return "neutral";
}

export function StatusIndicator({
  status,
  label,
  tone,
  className,
  labelClassName,
  ...props
}: Omit<ComponentProps<"span">, "children"> & {
  status: string;
  label?: string;
  tone?: StatusIndicatorTone;
  labelClassName?: string;
}) {
  const resolvedTone = tone ?? resolveStatusTone(status);
  const displayLabel = label ?? status.replaceAll("-", " ");

  return (
    <span
      data-slot="status-indicator"
      data-tone={resolvedTone}
      className={cn(statusIndicatorVariants({ tone: resolvedTone }), className)}
      {...props}
    >
      <span
        aria-hidden="true"
        className="inline-flex size-4 shrink-0 items-center justify-center"
        data-slot="status-indicator-glyph"
      >
        <span
          className="size-2 rounded-full bg-current"
          data-slot="status-indicator-dot"
        />
      </span>
      <span
        className={cn(
          "text-sm font-medium whitespace-nowrap capitalize",
          labelClassName
        )}
      >
        {displayLabel}
      </span>
    </span>
  );
}

export { statusIndicatorVariants };
