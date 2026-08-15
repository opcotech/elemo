import { Rows3Icon } from "lucide-react";

import { PriorityRibbon } from "./priority-ribbon";
import { dateLabel, workItemPath } from "./utils";
import { WorkLabelBadges } from "./work-label-badges";

import { AppList } from "@/components/shared/entity-link";
import { EmptyState } from "@/components/ui/empty-state";
import { InternalLink } from "@/components/ui/internal-link";
import { PersonAvatarStack } from "@/components/ui/person-avatar-stack";
import { StatusIndicator } from "@/components/ui/status-indicator";
import { internalPath } from "@/lib/internal-url";
import type { WorkItem } from "@/lib/mock-data";
import { cn } from "@/lib/utils";
import { workItemAssignmentPeople } from "@/lib/work/resolve-work-people";

export function CompactWorkList({
  items,
  onSelect,
  compact = false,
  emptyTitle = "No work here",
  emptyDescription = "Work matching this context will appear here.",
  limit,
  showAssignees = true,
  showLabels = true,
}: {
  items: readonly WorkItem[];
  onSelect?: (item: WorkItem) => void;
  compact?: boolean;
  emptyTitle?: string;
  emptyDescription?: string;
  limit?: number;
  showAssignees?: boolean;
  showLabels?: boolean;
}) {
  const visible = limit ? items.slice(0, limit) : items;

  if (visible.length === 0) {
    return (
      <EmptyState
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
            to={internalPath(workItemPath(item))}
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
          <div className="ml-auto flex shrink-0 items-center gap-6 sm:gap-8">
            <span className="hidden items-center sm:inline-flex">
              <StatusIndicator status={item.status} />
            </span>
            <span className="hidden shrink-0 items-center md:inline-flex">
              <PriorityRibbon priority={item.priority} />
            </span>
            {showAssignees ? (
              <span className="hidden shrink-0 items-center lg:inline-flex">
                <PersonAvatarStack
                  people={workItemAssignmentPeople(item)}
                  size="sm"
                />
              </span>
            ) : null}
            {showLabels ? (
              <WorkLabelBadges
                labelIds={item.labelIds}
                labels={item.labels}
                limit={2}
                className="max-w-32 shrink-0 flex-nowrap sm:max-w-40"
              />
            ) : null}
            <span className="text-muted-foreground hidden justify-end text-xs xl:inline-flex xl:items-center">
              {dateLabel(item.startDate)}
            </span>
            <span className="text-muted-foreground hidden justify-end text-xs xl:inline-flex xl:items-center">
              {dateLabel(item.dueDate)}
            </span>
          </div>
        </div>
      ))}
    </AppList>
  );
}
