import type { LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";

export function PropertyRibbon({
  icon: Icon,
  label,
  className,
  iconClassName,
  labelClassName,
  showLabel = true,
  "data-slot": dataSlot,
  "data-kind": dataKind,
  "data-priority": dataPriority,
}: {
  icon: LucideIcon;
  label: string;
  className?: string;
  iconClassName?: string;
  labelClassName?: string;
  showLabel?: boolean;
  "data-slot": string;
  "data-kind"?: string;
  "data-priority"?: string;
}) {
  return (
    <span
      className={cn("inline-flex items-center gap-2", className)}
      data-slot={dataSlot}
      data-kind={dataKind}
      data-priority={dataPriority}
    >
      <Icon
        aria-hidden="true"
        className={cn("size-4 shrink-0", iconClassName)}
      />
      {showLabel ? (
        <span className={cn("text-sm font-medium capitalize", labelClassName)}>
          {label}
        </span>
      ) : null}
    </span>
  );
}
