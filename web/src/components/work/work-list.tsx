import { Rows3Icon } from "lucide-react";

import { dateLabel } from "./utils";

import { AppEmptyState } from "@/components/shared/app-feedback";
import { AppList } from "@/components/shared/entity-link";
import { StatusIndicator } from "@/components/shared/status-indicator";
import { InternalLink } from "@/components/ui/internal-link";
import { internalPath } from "@/lib/internal-url";
import type { WorkItem } from "@/lib/mock-data";
import { cn } from "@/lib/utils";

export function CompactWorkList({
  items,
  onSelect,
  compact = false,
  emptyTitle = "No work here",
  emptyDescription = "Work matching this context will appear here.",
  limit,
}: {
  items: readonly WorkItem[];
  onSelect?: (item: WorkItem) => void;
  compact?: boolean;
  emptyTitle?: string;
  emptyDescription?: string;
  limit?: number;
}) {
  const visible = limit ? items.slice(0, limit) : items;

  if (visible.length === 0) {
    return (
      <AppEmptyState
        compact
        icon={<Rows3Icon />}
        title={emptyTitle}
        description={emptyDescription}
      />
    );
  }

  return (
    <AppList>
      {visible.map((item) => (
        <div
          key={item.id}
          role="listitem"
          className={cn(
            "group hover:bg-muted/50 flex min-w-0 items-center gap-3 px-3",
            compact ? "py-2" : "py-2.5"
          )}
        >
          <span
            aria-hidden
            className={cn(
              "size-2 shrink-0 rounded-full",
              item.status === "blocked"
                ? "bg-destructive"
                : item.status === "done"
                  ? "bg-success"
                  : "bg-primary"
            )}
          />
          <InternalLink
            to={internalPath(`/work/${item.id}`)}
            onClick={(event) => event.stopPropagation()}
            className="text-muted-foreground hover:text-primary w-20 shrink-0 font-mono text-xs"
          >
            {item.key}
          </InternalLink>
          {onSelect ? (
            <button
              type="button"
              className="hover:text-primary focus-visible:ring-ring min-w-0 flex-1 truncate rounded-sm text-left text-sm font-medium outline-none focus-visible:ring-2"
              aria-label={`Inspect ${item.key}: ${item.title}`}
              onClick={() => onSelect(item)}
            >
              {item.title}
            </button>
          ) : (
            <span className="min-w-0 flex-1 truncate text-sm font-medium">
              {item.title}
            </span>
          )}
          <span className="hidden sm:block">
            <StatusIndicator status={item.status} />
          </span>
          <span className="text-muted-foreground hidden w-20 text-right text-xs md:block">
            {dateLabel(item.dueDate)}
          </span>
        </div>
      ))}
    </AppList>
  );
}
