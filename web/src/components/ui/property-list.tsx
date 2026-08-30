import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

export const propertyControlClassName =
  "border-transparent bg-transparent shadow-none hover:bg-primary/5 hover:border-primary/20 hover:text-foreground aria-expanded:text-foreground dark:bg-transparent dark:hover:bg-primary/10 dark:hover:border-primary/30 h-8 w-full justify-between px-2 text-sm font-medium";

export function PropertyList({
  items,
  compact = false,
}: {
  items: readonly {
    id?: string;
    label: string;
    value: ReactNode;
    icon?: ReactNode;
  }[];
  compact?: boolean;
}) {
  return (
    <dl className="divide-border/60 divide-y">
      {items.map((item) => (
        <div
          key={item.id ?? item.label}
          className={cn(
            "grid grid-cols-[minmax(7rem,0.75fr)_minmax(0,1.25fr)] items-start gap-4",
            compact ? "py-2" : "py-3"
          )}
        >
          <dt className="text-muted-foreground flex items-center gap-2 text-xs">
            {item.icon}
            {item.label}
          </dt>
          <dd className="min-w-0 overflow-visible text-sm font-medium">
            {item.value}
          </dd>
        </div>
      ))}
    </dl>
  );
}
