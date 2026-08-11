import type { TimelineEntry } from "@/lib/mock-data";
import type { WorkRouteSearch } from "@/lib/work-route-search";

const DAY_IN_MS = 24 * 60 * 60 * 1000;
const WEEK_IN_MS = 7 * DAY_IN_MS;
const MINIMUM_TIMELINE_WEEKS = 4;

export type SearchPatch = Partial<WorkRouteSearch>;

export interface TimelineScale {
  start: number;
  end: number;
  ticks: readonly Date[];
}

export function dateLabel(value: string | null) {
  if (!value) return "Unscheduled";
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  }).format(new Date(value));
}

export function selectedWorkId(selected: string | undefined) {
  return selected?.startsWith("work:") ? selected.slice(5) : undefined;
}

export function parseSort(sort: string): WorkRouteSearch["sort"] {
  return sort.includes(":") ? sort : "rank:asc";
}

function startOfUtcWeek(timestamp: number) {
  const date = new Date(timestamp);
  const day = date.getUTCDay() || 7;
  date.setUTCDate(date.getUTCDate() - day + 1);
  date.setUTCHours(0, 0, 0, 0);
  return date.getTime();
}

function endOfUtcWeek(timestamp: number) {
  const start = startOfUtcWeek(timestamp);
  return timestamp === start ? start : start + WEEK_IN_MS;
}

export function createTimelineScale(
  entries: readonly TimelineEntry[]
): TimelineScale | undefined {
  if (entries.length === 0) return undefined;

  const firstStart = Math.min(
    ...entries.map((entry) => new Date(entry.startAt).getTime())
  );
  const lastEnd = Math.max(
    ...entries.map((entry) => new Date(entry.endAt).getTime())
  );
  const start = startOfUtcWeek(firstStart);
  const end = Math.max(
    endOfUtcWeek(lastEnd),
    start + MINIMUM_TIMELINE_WEEKS * WEEK_IN_MS
  );
  const ticks: Date[] = [];

  for (let time = start; time < end; time += WEEK_IN_MS) {
    ticks.push(new Date(time));
  }

  return { start, end, ticks };
}

export function timelinePosition(entry: TimelineEntry, scale: TimelineScale) {
  const duration = scale.end - scale.start;
  const entryStart = new Date(entry.startAt).getTime();
  const entryEnd = new Date(entry.endAt).getTime();
  const left = Math.max(
    0,
    Math.min(100, ((entryStart - scale.start) / duration) * 100)
  );
  const right = Math.max(
    left,
    Math.min(100, ((entryEnd - scale.start) / duration) * 100)
  );

  return {
    left,
    width: entry.kind === "milestone" ? 0 : Math.max(1, right - left),
  };
}

export function paginate<T>(
  items: readonly T[],
  page: number,
  pageSize: number
) {
  const pageCount = Math.max(1, Math.ceil(items.length / pageSize));
  const safePage = Math.min(Math.max(1, page), pageCount);
  const start = (safePage - 1) * pageSize;

  return {
    items: items.slice(start, start + pageSize),
    page: safePage,
    pageCount,
  };
}
