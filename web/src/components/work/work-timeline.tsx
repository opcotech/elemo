import { createTimelineScale, dateLabel, timelinePosition } from "./utils";

import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { getWorkItem, selectTimeline } from "@/lib/mock-data";
import type { Scope, WorkItem } from "@/lib/mock-data";
import { cn } from "@/lib/utils";

export function WorkTimeline({
  items,
  scope,
  compact,
  onSelect,
}: {
  items: readonly WorkItem[];
  scope: Scope;
  compact: boolean;
  onSelect: (item: WorkItem) => void;
}) {
  const itemIds = new Set(items.map((item) => item.id));
  const timeline = selectTimeline({ scope }).filter((entry) =>
    itemIds.has(entry.workItemId)
  );
  const scale = createTimelineScale(timeline);
  const datedIds = new Set(timeline.map((entry) => entry.workItemId));
  const unscheduled = items.filter((item) => !datedIds.has(item.id));
  const columns = scale?.ticks.length ?? 1;

  return (
    <ScrollArea className="h-full min-h-0 min-w-0">
      <div className="min-w-225">
        <div className="bg-background sticky top-0 z-10 grid grid-cols-[300px_1fr] border-b">
          <div className="border-r px-4 py-3 text-xs font-semibold uppercase">
            Work
          </div>
          <div
            className="grid px-4 py-3 text-xs"
            style={{
              gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))`,
            }}
          >
            {scale ? (
              scale.ticks.map((tick) => (
                <span key={tick.toISOString()}>
                  {dateLabel(tick.toISOString())}
                </span>
              ))
            ) : (
              <span>No scheduled dates</span>
            )}
          </div>
        </div>
        {scale &&
          timeline.map((entry) => {
            const item = getWorkItem(entry.workItemId);
            if (!item) return null;
            const position = timelinePosition(entry, scale);

            return (
              <button
                type="button"
                key={entry.id}
                onClick={() => onSelect(item)}
                className="hover:bg-muted/30 focus-visible:ring-ring grid w-full grid-cols-[300px_1fr] border-b text-left outline-none focus-visible:ring-2 focus-visible:ring-inset"
                aria-label={`Inspect ${item.key}: ${item.title}, ${dateLabel(entry.startAt)} to ${dateLabel(entry.endAt)}`}
              >
                <span
                  className={cn(
                    "truncate border-r px-4 text-sm",
                    compact ? "py-2" : "py-3"
                  )}
                >
                  <span className="text-muted-foreground mr-2 font-mono text-xs">
                    {item.key}
                  </span>
                  {item.title}
                </span>
                <span
                  className={cn(
                    "relative",
                    compact ? "my-2 h-5" : "my-2.5 h-6"
                  )}
                  style={{
                    backgroundImage:
                      "linear-gradient(to right, var(--border) 1px, transparent 1px)",
                    backgroundSize: `${100 / columns}% 100%`,
                  }}
                >
                  <span
                    aria-hidden
                    className={cn(
                      "bg-primary/80 absolute top-1 h-4 rounded-md",
                      entry.kind === "milestone" &&
                        "size-3 -translate-x-1/2 rotate-45 rounded-sm"
                    )}
                    style={{
                      left: `${position.left}%`,
                      width:
                        entry.kind === "milestone"
                          ? undefined
                          : `${position.width}%`,
                    }}
                  />
                </span>
              </button>
            );
          })}
        {unscheduled.length > 0 && (
          <div className="grid grid-cols-[300px_1fr] border-b">
            <div className="border-r px-4 py-3">
              <p className="text-xs font-semibold uppercase">
                Unscheduled ({unscheduled.length})
              </p>
              <div className="mt-2 max-h-32 space-y-1 overflow-y-auto">
                {unscheduled.map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    className="text-muted-foreground hover:text-primary focus-visible:ring-ring block w-full truncate rounded-sm text-left text-xs outline-none focus-visible:ring-2"
                    onClick={() => onSelect(item)}
                  >
                    {item.key} {item.title}
                  </button>
                ))}
              </div>
            </div>
            <div
              style={{
                backgroundImage:
                  "linear-gradient(to right, var(--border) 1px, transparent 1px)",
                backgroundSize: `${100 / columns}% 100%`,
              }}
            />
          </div>
        )}
      </div>
      <ScrollBar orientation="horizontal" />
    </ScrollArea>
  );
}
