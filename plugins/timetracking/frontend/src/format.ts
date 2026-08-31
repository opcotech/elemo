export function formatElapsed(seconds: number): string {
  const value = Math.max(0, Math.floor(seconds));
  const h = Math.floor(value / 3600);
  const m = Math.floor((value % 3600) / 60);
  const s = value % 60;
  return `${h}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

export function formatDate(value?: string | null): string {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function dateKey(value?: string | null): string {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value.slice(0, 10);
  }
  return date.toISOString().slice(0, 10);
}

export function secondsFromHoursMinutes(hours: number, minutes: number): number {
  const total = Math.floor(hours) * 3600 + Math.floor(minutes) * 60;
  return Math.max(1, total);
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

export function personName(user: {
  first_name: string;
  last_name: string;
  id: string;
}): string {
  const name = `${user.first_name} ${user.last_name}`.trim();
  return name || user.id;
}
