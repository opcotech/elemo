import { PlusIcon } from "lucide-react";
import { useId, useState } from "react";

import { WorkCard } from "./work-card";

import { openQuickCreate } from "@/components/quick-create/open";
import { Button } from "@/components/ui/button";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { getPerson } from "@/lib/mock-data";
import type { WorkItem, WorkStatus } from "@/lib/mock-data";
import type { WorkRouteSearch } from "@/lib/work-route-search";

const COLUMN_PAGE_SIZE = 25;
const statusOrder: readonly WorkStatus[] = [
  "backlog",
  "planned",
  "in-progress",
  "blocked",
  "done",
  "canceled",
];

function BoardColumn({
  label,
  items,
  compact,
  onSelect,
}: {
  label: string;
  items: readonly WorkItem[];
  compact: boolean;
  onSelect: (item: WorkItem) => void;
}) {
  const headingId = useId();
  const [visibleCount, setVisibleCount] = useState(COLUMN_PAGE_SIZE);
  const visibleItems = items.slice(0, visibleCount);
  const remaining = items.length - visibleItems.length;

  return (
    <section
      aria-labelledby={headingId}
      className="bg-surface-sunken flex max-h-[calc(100svh-15rem)] w-72 flex-col rounded-xl border"
    >
      <header className="flex shrink-0 items-center gap-2 border-b px-3 py-2.5">
        <h2
          id={headingId}
          className="flex-1 text-xs font-semibold tracking-wide uppercase"
        >
          {label.replaceAll("-", " ")}
        </h2>
        <span className="text-muted-foreground text-xs tabular-nums">
          {items.length}
        </span>
        <Button
          variant="ghost"
          size="icon-xs"
          onClick={() => openQuickCreate("work")}
          aria-label={`Add work to ${label}`}
        >
          <PlusIcon />
        </Button>
      </header>
      <div className="min-h-20 space-y-2 overflow-y-auto p-2">
        {visibleItems.map((item) => (
          <WorkCard
            key={item.id}
            item={item}
            compact={compact}
            onSelect={onSelect}
          />
        ))}
        {items.length === 0 && (
          <p className="text-muted-foreground py-6 text-center text-xs">
            No work in this group
          </p>
        )}
        {remaining > 0 && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="w-full"
            onClick={() => setVisibleCount((count) => count + COLUMN_PAGE_SIZE)}
          >
            Show {Math.min(remaining, COLUMN_PAGE_SIZE)} more
          </Button>
        )}
      </div>
    </section>
  );
}

export function WorkBoard({
  items,
  group,
  compact,
  onSelect,
}: {
  items: readonly WorkItem[];
  group: WorkRouteSearch["group"];
  compact: boolean;
  onSelect: (item: WorkItem) => void;
}) {
  const groupValue = (item: WorkItem) => {
    if (group === "priority") return item.priority;
    if (group === "assignee") {
      return item.assigneeId
        ? (getPerson(item.assigneeId)?.displayName ?? "Assigned")
        : "Unassigned";
    }
    if (group === "none") return "All work";
    return item.status;
  };
  const grouped = Map.groupBy(items, groupValue);
  const keys =
    group === "status"
      ? statusOrder
      : ([...grouped.keys()] as readonly string[]);

  return (
    <ScrollArea className="h-full min-h-0 min-w-0">
      <div className="flex min-w-max items-start gap-3">
        {keys.map((key) => (
          <BoardColumn
            key={key}
            label={key}
            items={grouped.get(key) ?? []}
            compact={compact}
            onSelect={onSelect}
          />
        ))}
      </div>
      <ScrollBar orientation="horizontal" />
    </ScrollArea>
  );
}
