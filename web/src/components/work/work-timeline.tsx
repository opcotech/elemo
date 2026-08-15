import { useEffect, useRef, useState } from "react";
import type {
  CSSProperties,
  MouseEvent as ReactMouseEvent,
  PointerEvent as ReactPointerEvent,
  RefObject,
} from "react";

import type { TimelineDateChange } from "./use-timeline-issue-dates";
import {
  applyTimelineDrag,
  createTimelineScale,
  dateLabel,
  timelinePosition,
  timelineRangeFromSpan,
  timelineRangeLabel,
  utcDayDelta,
  workItemDatesFromTimelineRange,
  workItemsToTimelineEntries,
} from "./utils";
import type {
  TimelineDragMode,
  TimelineScale,
  WorkTimelineEntry,
  WorkTimelineRange,
} from "./utils";

import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { WorkItem } from "@/lib/mock-data";
import { cn } from "@/lib/utils";

const DRAG_THRESHOLD_PX = 4;
const WORK_COLUMN_PX = 300;
const WEEK_COLUMN_MIN = "16rem";

interface TimelineDrag {
  readonly mode: TimelineDragMode;
  readonly originX: number;
  readonly startAt: string;
  readonly endAt: string;
  readonly trackWidth: number;
  readonly pointerId: number;
}

export function WorkTimeline({
  items,
  compact,
  onSelect,
  onDatesChange,
}: {
  items: readonly WorkItem[];
  compact: boolean;
  onSelect: (item: WorkItem) => void;
  onDatesChange: (change: TimelineDateChange) => void;
}) {
  const timeline = workItemsToTimelineEntries(items);
  const scale = createTimelineScale(timeline);
  const datedIds = new Set(timeline.map((entry) => entry.item.id));
  const unscheduled = items.filter((item) => !datedIds.has(item.id));
  const columns = scale?.ticks.length ?? 1;
  const scaleStart = scale?.start;
  const scaleEnd = scale?.end;
  const rootRef = useRef<HTMLDivElement>(null);
  const trackColumnsStyle = {
    gridTemplateColumns: `repeat(${columns}, minmax(${WEEK_COLUMN_MIN}, 1fr))`,
  };
  const trackGridStyle = {
    ...trackColumnsStyle,
    backgroundImage:
      "linear-gradient(to right, var(--border) 1px, transparent 1px)",
    backgroundSize: `calc(100% / ${columns}) 100%`,
  };

  useEffect(() => {
    if (scaleStart == null || scaleEnd == null) {
      return;
    }
    const viewport = rootRef.current?.querySelector(
      '[data-slot="scroll-area-viewport"]'
    );
    if (!(viewport instanceof HTMLElement)) {
      return;
    }

    const now = Date.now();
    const maxScroll = viewport.scrollWidth - viewport.clientWidth;
    if (maxScroll <= 0) {
      return;
    }
    if (now <= scaleStart) {
      viewport.scrollLeft = 0;
      return;
    }
    if (now >= scaleEnd) {
      viewport.scrollLeft = maxScroll;
      return;
    }

    const trackWidth = Math.max(0, viewport.scrollWidth - WORK_COLUMN_PX);
    const offset = ((now - scaleStart) / (scaleEnd - scaleStart)) * trackWidth;
    viewport.scrollLeft = Math.min(maxScroll, Math.max(0, offset));
  }, [scaleEnd, scaleStart]);

  return (
    <div ref={rootRef} className="h-full min-h-0 min-w-0">
      <ScrollArea className="h-full min-h-0 min-w-0">
        <TooltipProvider delay={300}>
          <div
            style={{
              minWidth: `calc(${WORK_COLUMN_PX}px + ${columns} * ${WEEK_COLUMN_MIN})`,
            }}
          >
            <div
              className="bg-background sticky top-0 z-20 grid border-b"
              style={{
                gridTemplateColumns: `${WORK_COLUMN_PX}px minmax(0, 1fr)`,
              }}
            >
              <div className="bg-background sticky left-0 z-30 border-r px-4 py-3 text-xs font-semibold uppercase">
                Work
              </div>
              <div className="grid py-3 text-xs" style={trackColumnsStyle}>
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
              timeline.map((entry) => (
                <TimelineRow
                  key={entry.item.id}
                  entry={entry}
                  scale={scale}
                  compact={compact}
                  trackGridStyle={trackGridStyle}
                  onSelect={onSelect}
                  onDatesChange={onDatesChange}
                />
              ))}
            {unscheduled.length > 0 && (
              <div
                className="grid border-b"
                style={{
                  gridTemplateColumns: `${WORK_COLUMN_PX}px minmax(0, 1fr)`,
                }}
              >
                <div className="bg-background sticky left-0 z-10 border-r px-4 py-3">
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
                <div style={trackGridStyle} />
              </div>
            )}
          </div>
        </TooltipProvider>
        <ScrollBar orientation="horizontal" />
      </ScrollArea>
    </div>
  );
}

function TimelineRow({
  entry,
  scale,
  compact,
  trackGridStyle,
  onSelect,
  onDatesChange,
}: {
  entry: WorkTimelineEntry;
  scale: TimelineScale;
  compact: boolean;
  trackGridStyle: CSSProperties;
  onSelect: (item: WorkItem) => void;
  onDatesChange: (change: TimelineDateChange) => void;
}) {
  const item = entry.item;
  const trackRef = useRef<HTMLDivElement>(null);
  const rangeLabel = timelineRangeLabel(entry);

  return (
    <div
      className="group bg-background grid w-full border-b"
      style={{
        gridTemplateColumns: `${WORK_COLUMN_PX}px minmax(0, 1fr)`,
      }}
    >
      <button
        type="button"
        onClick={() => onSelect(item)}
        className={cn(
          "bg-background focus-visible:ring-ring sticky left-0 z-50 truncate border-r px-4 text-left text-sm outline-none focus-visible:ring-2 focus-visible:ring-inset",
          compact ? "py-2" : "py-3",
          entry.kind === "milestone"
            ? "text-pink-700 hover:text-pink-800 dark:text-pink-400 dark:hover:text-pink-300"
            : "text-primary hover:text-primary-pressed"
        )}
        aria-label={`Inspect ${item.key}: ${item.title}, ${rangeLabel}`}
      >
        <span className="mr-2 font-mono text-xs">{item.key}</span>
        {item.title}
      </button>
      <div
        ref={trackRef}
        className={cn("relative", compact ? "my-2 h-5" : "my-2.5 h-6")}
        style={trackGridStyle}
      >
        <TimelineBar
          entry={entry}
          scale={scale}
          trackRef={trackRef}
          onSelect={onSelect}
          onDatesChange={onDatesChange}
        />
      </div>
    </div>
  );
}

function TimelineBar({
  entry,
  scale,
  trackRef,
  onSelect,
  onDatesChange,
}: {
  entry: WorkTimelineEntry;
  scale: TimelineScale;
  trackRef: RefObject<HTMLDivElement | null>;
  onSelect: (item: WorkItem) => void;
  onDatesChange: (change: TimelineDateChange) => void;
}) {
  const item = entry.item;
  const [drag, setDrag] = useState<TimelineDrag | null>(null);
  const [preview, setPreview] = useState<WorkTimelineRange | null>(null);
  const previewRef = useRef(timelineRangeFromSpan(entry.startAt, entry.endAt));
  const movedRef = useRef(false);
  const suppressClickRef = useRef(false);

  const [tooltipOpen, setTooltipOpen] = useState(false);
  const range = preview ?? timelineRangeFromSpan(entry.startAt, entry.endAt);
  const position = timelinePosition(range, scale);
  const rangeLabel = timelineRangeLabel(range);
  const dragging = drag !== null;

  useEffect(() => {
    previewRef.current = timelineRangeFromSpan(entry.startAt, entry.endAt);
    setPreview(null);
  }, [entry.startAt, entry.endAt]);

  useEffect(() => {
    if (!drag) {
      return;
    }

    const onMove = (event: PointerEvent) => {
      if (event.pointerId !== drag.pointerId) {
        return;
      }
      event.preventDefault();
      const deltaPx = event.clientX - drag.originX;
      if (Math.abs(deltaPx) >= DRAG_THRESHOLD_PX) {
        movedRef.current = true;
      }
      const days = utcDayDelta(deltaPx, drag.trackWidth, scale);
      const next = applyTimelineDrag({
        startAt: drag.startAt,
        endAt: drag.endAt,
        mode: drag.mode,
        days,
      });
      const nextRange = timelineRangeFromSpan(next.startAt, next.endAt);
      previewRef.current = nextRange;
      setPreview(nextRange);
    };

    const onUp = (event: PointerEvent) => {
      if (event.pointerId !== drag.pointerId) {
        return;
      }
      const nextRange = previewRef.current;
      const moved = movedRef.current;
      movedRef.current = false;
      setDrag(null);
      if (!moved) {
        setPreview(null);
        return;
      }
      suppressClickRef.current = true;
      const dates = workItemDatesFromTimelineRange(item, nextRange);
      if (
        dates.startDate === item.startDate &&
        dates.dueDate === item.dueDate
      ) {
        setPreview(null);
        return;
      }
      onDatesChange({ item, ...dates });
    };

    window.addEventListener("pointermove", onMove, { passive: false });
    window.addEventListener("pointerup", onUp);
    window.addEventListener("pointercancel", onUp);
    return () => {
      window.removeEventListener("pointermove", onMove);
      window.removeEventListener("pointerup", onUp);
      window.removeEventListener("pointercancel", onUp);
    };
  }, [drag, item, onDatesChange, scale]);

  const startDrag = (
    event: ReactPointerEvent<HTMLElement>,
    mode: TimelineDragMode
  ) => {
    if (event.button !== 0) {
      return;
    }
    event.stopPropagation();
    event.currentTarget.setPointerCapture(event.pointerId);
    movedRef.current = false;
    setTooltipOpen(false);
    previewRef.current = timelineRangeFromSpan(entry.startAt, entry.endAt);
    setPreview(previewRef.current);
    setDrag({
      mode,
      originX: event.clientX,
      startAt: entry.startAt,
      endAt: entry.endAt,
      trackWidth: trackRef.current?.getBoundingClientRect().width ?? 0,
      pointerId: event.pointerId,
    });
  };

  const onBarClick = (event: ReactMouseEvent<HTMLElement>) => {
    if (suppressClickRef.current) {
      suppressClickRef.current = false;
      event.preventDefault();
      event.stopPropagation();
      return;
    }
    onSelect(item);
  };

  return (
    <Tooltip
      open={dragging || tooltipOpen}
      onOpenChange={(open) => {
        if (!dragging) {
          setTooltipOpen(open);
        }
      }}
    >
      <TooltipTrigger
        render={
          <span
            className={cn(
              "group absolute top-1 h-4 touch-none select-none",
              range.kind === "milestone" && "w-3 -translate-x-1/2",
              dragging ? "z-10 cursor-grabbing" : "cursor-grab"
            )}
            style={{
              left: `${position.left}%`,
              width:
                range.kind === "milestone" ? undefined : `${position.width}%`,
            }}
            onPointerDown={(event) => startDrag(event, "move")}
            onClick={onBarClick}
          />
        }
      >
        <span
          className={cn(
            "bg-primary/20 border-primary/80 pointer-events-none absolute inset-0 rounded-sm border",
            range.kind === "milestone" && "border-pink-500/80 bg-pink-500/20"
          )}
        />
        <TimelineHandle
          edge="start"
          dragging={dragging}
          onPointerDown={(event) => startDrag(event, "start")}
        />
        <TimelineHandle
          edge="end"
          dragging={dragging}
          onPointerDown={(event) => startDrag(event, "end")}
        />
      </TooltipTrigger>
      <TooltipContent>{rangeLabel}</TooltipContent>
    </Tooltip>
  );
}

function TimelineHandle({
  edge,
  dragging,
  onPointerDown,
}: {
  edge: "start" | "end";
  dragging: boolean;
  onPointerDown: (event: ReactPointerEvent<HTMLButtonElement>) => void;
}) {
  return (
    <button
      type="button"
      aria-label={edge === "start" ? "Adjust start date" : "Adjust due date"}
      className={cn(
        "border-primary bg-background absolute top-0 z-10 h-full w-1.5 rounded-sm border opacity-0 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100",
        edge === "start"
          ? "left-0 -translate-x-1/2"
          : "right-0 translate-x-1/2",
        dragging && "opacity-100",
        "cursor-ew-resize touch-none"
      )}
      onPointerDown={(event) => {
        event.stopPropagation();
        onPointerDown(event);
      }}
      onClick={(event) => event.stopPropagation()}
    />
  );
}
