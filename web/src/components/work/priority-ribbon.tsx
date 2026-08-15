import {
  ChevronDownIcon,
  ChevronUpIcon,
  ChevronsDownIcon,
  ChevronsUpIcon,
  EqualIcon,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { PropertyRibbon } from "./property-ribbon";

import type { IssuePriority } from "@/lib/api/types";
import { cn } from "@/lib/utils";

const priorityIcons: Record<IssuePriority, LucideIcon> = {
  highest: ChevronsUpIcon,
  high: ChevronUpIcon,
  normal: EqualIcon,
  low: ChevronDownIcon,
  lowest: ChevronsDownIcon,
};

const priorityToneClassName: Record<IssuePriority, string> = {
  highest: "text-destructive",
  high: "text-warning",
  normal: "text-foreground",
  low: "text-primary",
  lowest: "text-muted-foreground",
};

const priorityLabels: Record<IssuePriority, string> = {
  highest: "Highest",
  high: "High",
  normal: "Normal",
  low: "Low",
  lowest: "Lowest",
};

export function PriorityRibbon({
  priority,
  className,
  labelClassName,
  showLabel = true,
}: {
  priority: IssuePriority;
  className?: string;
  labelClassName?: string;
  showLabel?: boolean;
}) {
  return (
    <PropertyRibbon
      icon={priorityIcons[priority]}
      label={priorityLabels[priority]}
      className={cn(priorityToneClassName[priority], className)}
      labelClassName={labelClassName}
      showLabel={showLabel}
      data-slot="priority-ribbon"
      data-priority={priority}
    />
  );
}

export { priorityLabels as issuePriorityLabels };
