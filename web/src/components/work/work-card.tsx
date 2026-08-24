import { PriorityRibbon } from "./priority-ribbon";
import { dateLabel, workItemPath } from "./utils";
import { WorkLabelBadges } from "./work-label-badges";

import { InternalLink } from "@/components/ui/internal-link";
import { PersonAvatarStack } from "@/components/ui/person-avatar-stack";
import { StatusIndicator } from "@/components/ui/status-indicator";
import { internalPath } from "@/lib/internal-url";
import { cn } from "@/lib/utils";
import type { WorkItem } from "@/lib/work/model";
import { workItemAssignmentPeople } from "@/lib/work/resolve-work-people";

function WorkIdentity({
  item,
  compact,
  onSelect,
}: {
  item: WorkItem;
  compact?: boolean;
  onSelect?: (item: WorkItem) => void;
}) {
  const people = workItemAssignmentPeople(item);

  return (
    <div className="min-w-0 flex-1">
      <div className="text-muted-foreground flex items-center gap-2 text-xs font-medium">
        <InternalLink
          to={internalPath(workItemPath(item))}
          className="hover:text-primary focus-visible:ring-ring focus-visible:rounded focus-visible:ring-2 focus-visible:outline-none"
          onClick={(event) => event.stopPropagation()}
        >
          {item.key}
        </InternalLink>
        <StatusIndicator labelClassName="text-xs" status={item.status} />
      </div>
      {onSelect ? (
        <button
          type="button"
          className={cn(
            "hover:text-primary focus-visible:ring-ring mt-1 block w-full rounded-sm text-left font-medium outline-none focus-visible:ring-2",
            compact ? "text-sm" : "text-[15px]"
          )}
          aria-label={`Inspect ${item.key}: ${item.title}`}
          onClick={() => onSelect(item)}
        >
          {item.title}
        </button>
      ) : (
        <p
          className={cn(
            "mt-1 font-medium",
            compact ? "text-sm" : "text-[15px]"
          )}
        >
          {item.title}
        </p>
      )}
      {!compact && (
        <p className="text-muted-foreground mt-1 line-clamp-2 text-xs whitespace-pre-wrap">
          {item.summary}
        </p>
      )}
      <WorkLabelBadges
        labelIds={item.labelIds}
        labels={item.labels}
        limit={compact ? 2 : 3}
        className="mt-2"
      />
      <div className="text-muted-foreground mt-4 flex flex-wrap items-center gap-3 text-xs">
        <PersonAvatarStack people={people} size="sm" />
        <PriorityRibbon labelClassName="text-xs" priority={item.priority} />
        {item.dueDate && <span>{dateLabel(item.dueDate)}</span>}
      </div>
    </div>
  );
}

export function WorkCard({
  item,
  compact = false,
  onSelect,
  className,
}: {
  item: WorkItem;
  compact?: boolean;
  onSelect?: (item: WorkItem) => void;
  className?: string;
}) {
  return (
    <article
      className={cn(
        "bg-card hover:border-primary/30 w-full rounded-lg border text-left shadow-xs transition-colors",
        compact ? "p-2.5" : "p-3",
        className
      )}
    >
      <WorkIdentity item={item} compact={compact} onSelect={onSelect} />
    </article>
  );
}
