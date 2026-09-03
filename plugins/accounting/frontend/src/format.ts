export function formatHours(seconds: number): string {
  const value = Math.max(0, Math.floor(seconds));
  const h = Math.floor(value / 3600);
  const m = Math.floor((value % 3600) / 60);
  if (h === 0) {
    return `${m}m`;
  }
  if (m === 0) {
    return `${h}h`;
  }
  return `${h}h ${m}m`;
}

export function secondsFromHoursMinutes(
  hours: number,
  minutes: number,
): number {
  return Math.max(0, Math.floor(hours) * 3600 + Math.floor(minutes) * 60);
}

export function hoursMinutesFromSeconds(seconds: number): {
  hours: number;
  minutes: number;
} {
  const value = Math.max(0, Math.floor(seconds));
  return {
    hours: Math.floor(value / 3600),
    minutes: Math.floor((value % 3600) / 60),
  };
}

export function parseISODate(value?: string | null): Date | null {
  if (!value) {
    return null;
  }
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value.slice(0, 10));
  if (!match) {
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? null : date;
  }
  const date = new Date(
    Number(match[1]),
    Number(match[2]) - 1,
    Number(match[3]),
  );
  return Number.isNaN(date.getTime()) ? null : date;
}

export function formatISODate(date: Date | null): string {
  if (!date) {
    return "";
  }
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function formatPeriod(
  start?: string | null,
  end?: string | null,
): string {
  const from = parseISODate(start);
  const to = parseISODate(end);
  if (!from && !to) {
    return "No period";
  }
  const opts: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "short",
    day: "numeric",
  };
  const left = from ? from.toLocaleDateString(undefined, opts) : "…";
  const right = to ? to.toLocaleDateString(undefined, opts) : "…";
  return `${left} – ${right}`;
}

export function dateKey(value?: string | null | Date): string {
  if (!value) {
    return "";
  }
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) {
    return typeof value === "string" ? value.slice(0, 10) : "";
  }
  return date.toISOString().slice(0, 10);
}

export function formatDateTime(value?: string | null): string {
  if (!value) {
    return "Unknown";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export function currentYearBounds(): { start: Date; end: Date } {
  const year = new Date().getFullYear();
  return {
    start: new Date(year, 0, 1),
    end: new Date(year, 11, 31),
  };
}
