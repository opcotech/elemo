import {
  todoPriorityIcons,
  todoPriorityLabels,
  todoPriorityToneClassName,
} from "./priority";

import { PropertyRibbon } from "@/components/work/property-ribbon";
import type { TodoPriority } from "@/lib/api/types";
import { cn } from "@/lib/utils";

export function TodoPriorityRibbon({
  priority,
  className,
  iconClassName,
  labelClassName,
  showLabel = true,
}: {
  priority: TodoPriority;
  className?: string;
  iconClassName?: string;
  labelClassName?: string;
  showLabel?: boolean;
}) {
  return (
    <PropertyRibbon
      icon={todoPriorityIcons[priority]}
      label={todoPriorityLabels[priority]}
      className={cn(todoPriorityToneClassName[priority], className)}
      iconClassName={iconClassName}
      labelClassName={labelClassName}
      showLabel={showLabel}
      data-slot="todo-priority-ribbon"
      data-priority={priority}
    />
  );
}
