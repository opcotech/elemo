import { format } from "date-fns";

import type { TimelineEntry } from "@/lib/mock-data";
import { workItemPath as canonicalWorkItemPath } from "@/lib/paths";
import type { WorkItem } from "@/lib/work/model";
import type { WorkRouteSearch } from "@/lib/work-route-search";

const DAY_IN_MS = 24 * 60 * 60 * 1000;
const WEEK_IN_MS = 7 * DAY_IN_MS;

export type SearchPatch = Partial<WorkRouteSearch>;

export interface TimelineScale {
  start: number;
  end: number;
  ticks: readonly Date[];
}

export interface WorkTimelineRange {
  readonly startAt: string;
  readonly endAt: string;
  readonly kind: TimelineEntry["kind"];
}

export interface WorkTimelineEntry extends WorkTimelineRange {
  readonly item: WorkItem;
}

/** Persist a DatePicker calendar day as UTC noon so date-only fields do not shift. */
export function calendarDateToUtcNoonIso(date: Date): string {
  return new Date(
    Date.UTC(date.getFullYear(), date.getMonth(), date.getDate(), 12, 0, 0, 0)
  ).toISOString();
}

/** Local midnight of the UTC calendar day, for DatePicker round-trip. */
export function utcIsoToCalendarDate(value: string): Date {
  const date = new Date(value);
  return new Date(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate());
}

export function dateLabel(value: string | null) {
  if (!value) return "Unscheduled";
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  }).format(new Date(value));
}

export function timelineRangeLabel(entry: WorkTimelineRange) {
  const from = dateLabel(entry.startAt);
  const to = dateLabel(entry.endAt);
  if (entry.kind === "milestone" || from === to) {
    return from;
  }
  return `${from} – ${to}`;
}

/** Full target date label matching DatePicker (`PPP`) on the UTC calendar day. */
export function formatTargetDate(value: string | null) {
  if (!value) return "Unscheduled";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Unscheduled";
  }
  return format(utcIsoToCalendarDate(value), "PPP");
}

/** Date and time label (`PPp`), or an em dash when missing/invalid. */
export function formatDateTime(value: string | null | undefined) {
  if (!value) {
    return "—";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "—";
  }
  return format(date, "PPp");
}

export function workItemPath(item: {
  readonly organizationSlug?: string | null;
  readonly namespaceSlug?: string | null;
  readonly namespace?: {
    readonly slug?: string;
    readonly organizationSlug?: string;
  };
  readonly key: string;
}): string {
  return canonicalWorkItemPath({
    organizationSlug:
      item.organizationSlug ?? item.namespace?.organizationSlug ?? "",
    namespaceSlug: item.namespaceSlug ?? item.namespace?.slug ?? "",
    issueKey: item.key,
  });
}

export function workItemUrl(item: {
  readonly organizationSlug?: string | null;
  readonly namespaceSlug?: string | null;
  readonly namespace?: {
    readonly slug?: string;
    readonly organizationSlug?: string;
  };
  readonly key: string;
}): string {
  const path = workItemPath(item);
  if (typeof window === "undefined") {
    return path;
  }
  return new URL(path, window.location.origin).href;
}

export function selectedWorkId(selected: string | undefined) {
  return selected?.startsWith("work:") ? selected.slice(5) : undefined;
}

export function parseSort(sort: string): WorkRouteSearch["sort"] {
  return sort.includes(":") ? sort : "rank:asc";
}

function startOfUtcDay(value: number) {
  const date = new Date(value);
  date.setUTCHours(0, 0, 0, 0);
  return date.getTime();
}

function startOfUtcWeek(value: number) {
  const date = new Date(value);
  const day = date.getUTCDay() || 7;
  date.setUTCDate(date.getUTCDate() - day + 1);
  date.setUTCHours(0, 0, 0, 0);
  return date.getTime();
}

function addUtcMonths(value: number, months: number) {
  const date = new Date(value);
  date.setUTCMonth(date.getUTCMonth() + months);
  return date.getTime();
}

function timestamp(value: string | null | undefined): number | null {
  if (!value) {
    return null;
  }
  const time = new Date(value).getTime();
  return Number.isNaN(time) ? null : time;
}

export function workItemTimelineRange(
  item: Pick<WorkItem, "startDate" | "dueDate">
): WorkTimelineRange | undefined {
  const startDate = item.startDate;
  const dueDate = item.dueDate;
  const startAt = timestamp(startDate);
  const endAt = timestamp(dueDate);

  if (startAt == null && endAt == null) {
    return undefined;
  }
  if (startAt == null && dueDate) {
    return { startAt: dueDate, endAt: dueDate, kind: "milestone" };
  }
  if (endAt == null && startDate) {
    return { startAt: startDate, endAt: startDate, kind: "milestone" };
  }
  if (startDate && dueDate && startAt != null && endAt != null) {
    if (startAt === endAt) {
      return { startAt: startDate, endAt: dueDate, kind: "milestone" };
    }
    if (startAt < endAt) {
      return { startAt: startDate, endAt: dueDate, kind: "work" };
    }
    return { startAt: dueDate, endAt: startDate, kind: "work" };
  }
  return undefined;
}

export function workItemsToTimelineEntries(
  items: readonly WorkItem[]
): WorkTimelineEntry[] {
  const entries: WorkTimelineEntry[] = [];

  for (const item of items) {
    const range = workItemTimelineRange(item);
    if (!range) {
      continue;
    }
    entries.push({ item, ...range });
  }

  return entries.toSorted(
    (left, right) =>
      left.startAt.localeCompare(right.startAt) ||
      left.item.id.localeCompare(right.item.id)
  );
}

export function createTimelineScale(
  entries: readonly WorkTimelineRange[]
): TimelineScale | undefined {
  if (entries.length === 0) return undefined;

  const firstStart = Math.min(
    ...entries.map((entry) => new Date(entry.startAt).getTime())
  );
  const lastEnd = Math.max(
    ...entries.map((entry) => new Date(entry.endAt).getTime())
  );
  const start = startOfUtcWeek(addUtcMonths(firstStart, -1));
  const end = startOfUtcWeek(addUtcMonths(lastEnd, 3)) + WEEK_IN_MS;
  const ticks: Date[] = [];

  for (let time = start; time < end; time += WEEK_IN_MS) {
    ticks.push(new Date(time));
  }

  return { start, end, ticks };
}

export function timelinePosition(
  entry: WorkTimelineRange,
  scale: TimelineScale
) {
  const duration = scale.end - scale.start;
  const entryStart = startOfUtcDay(new Date(entry.startAt).getTime());
  const entryEnd = startOfUtcDay(new Date(entry.endAt).getTime());
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
    right,
    width: entry.kind === "milestone" ? 0 : right - left,
  };
}

export type TimelineDragMode = "move" | "start" | "end";

export function utcDayDelta(
  deltaPx: number,
  trackWidth: number,
  scale: TimelineScale
) {
  if (trackWidth <= 0) {
    return 0;
  }
  const duration = scale.end - scale.start;
  return Math.round((deltaPx / trackWidth) * (duration / DAY_IN_MS));
}

export function shiftIsoByUtcDays(iso: string, days: number) {
  if (days === 0) {
    return iso;
  }
  const date = new Date(iso);
  date.setUTCDate(date.getUTCDate() + days);
  return date.toISOString();
}

export function timelineRangeFromSpan(
  startAt: string,
  endAt: string
): WorkTimelineRange {
  const startMs = new Date(startAt).getTime();
  const endMs = new Date(endAt).getTime();
  return {
    startAt,
    endAt,
    kind: startMs === endMs ? "milestone" : "work",
  };
}

export function applyTimelineDrag({
  startAt,
  endAt,
  mode,
  days,
}: {
  startAt: string;
  endAt: string;
  mode: TimelineDragMode;
  days: number;
}): { startAt: string; endAt: string } {
  if (days === 0) {
    return { startAt, endAt };
  }

  if (mode === "move") {
    return {
      startAt: shiftIsoByUtcDays(startAt, days),
      endAt: shiftIsoByUtcDays(endAt, days),
    };
  }

  if (mode === "start") {
    const nextStart = shiftIsoByUtcDays(startAt, days);
    if (new Date(nextStart).getTime() > new Date(endAt).getTime()) {
      return { startAt: endAt, endAt };
    }
    return { startAt: nextStart, endAt };
  }

  const nextEnd = shiftIsoByUtcDays(endAt, days);
  if (new Date(nextEnd).getTime() < new Date(startAt).getTime()) {
    return { startAt, endAt: startAt };
  }
  return { startAt, endAt: nextEnd };
}

export function workItemDatesFromTimelineRange(
  item: Pick<WorkItem, "startDate" | "dueDate">,
  range: Pick<WorkTimelineRange, "startAt" | "endAt">
): Pick<WorkItem, "startDate" | "dueDate"> {
  const isPoint =
    new Date(range.startAt).getTime() === new Date(range.endAt).getTime();
  if (isPoint) {
    if (item.startDate && !item.dueDate) {
      return { startDate: range.startAt, dueDate: null };
    }
    if (item.dueDate && !item.startDate) {
      return { startDate: null, dueDate: range.endAt };
    }
  }
  return { startDate: range.startAt, dueDate: range.endAt };
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
