import { UserIcon } from "lucide-react";

import { dateLabel } from "./utils";

import { StatusIndicator } from "@/components/shared/status-indicator";
import { InternalLink } from "@/components/ui/internal-link";
import { internalPath } from "@/lib/internal-url";
import { getPerson } from "@/lib/mock-data";
import type { WorkItem } from "@/lib/mock-data";
import { cn } from "@/lib/utils";

function WorkIdentity({
  item,
  compact,
  onSelect,
}: {
  item: WorkItem;
  compact?: boolean;
  onSelect?: (item: WorkItem) => void;
}) {
  const assignee = item.assigneeId ? getPerson(item.assigneeId) : undefined;

  return (
    <div className="min-w-0 flex-1">
      <div className="text-muted-foreground flex items-center gap-2 text-xs font-medium">
        <InternalLink
          to={internalPath(`/work/${item.id}`)}
          className="hover:text-primary focus-visible:ring-ring focus-visible:rounded focus-visible:ring-2 focus-visible:outline-none"
          onClick={(event) => event.stopPropagation()}
        >
          {item.key}
        </InternalLink>
        <span className="capitalize">{item.priority}</span>
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
        <p className="text-muted-foreground mt-1 line-clamp-2 text-xs leading-5">
          {item.summary}
        </p>
      )}
      <div className="text-muted-foreground mt-2 flex items-center gap-3 text-xs">
        <StatusIndicator status={item.status} />
        {assignee && (
          <span className="inline-flex items-center gap-1">
            <UserIcon className="size-3" />
            {assignee.displayName}
          </span>
        )}
        <span>{dateLabel(item.dueDate)}</span>
      </div>
    </div>
  );
}

export function WorkCard({
  item,
  compact = false,
  onSelect,
}: {
  item: WorkItem;
  compact?: boolean;
  onSelect?: (item: WorkItem) => void;
}) {
  return (
    <article
      className={cn(
        "bg-card hover:border-primary/30 w-full rounded-lg border text-left shadow-xs transition-colors",
        compact ? "p-2.5" : "p-3"
      )}
    >
      <WorkIdentity item={item} compact={compact} onSelect={onSelect} />
    </article>
  );
}
