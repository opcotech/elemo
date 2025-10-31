import { clsx } from "clsx";
import type { ClassValue } from "clsx";
import { format } from "date-fns";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function assertValue<T>(v: T | undefined, errorMessage: string): T {
  if (v === undefined) {
    throw new Error(errorMessage);
  }
  return v;
}

/**
 * Formats a date string to a human-readable format.
 *
 * @param dateString - The date string to format (ISO 8601 format or null)
 * @returns Formatted date string (e.g., "Jan 1, 2024") or "N/A" if invalid
 */
export function formatDate(dateString: string | null): string {
  if (!dateString) return "N/A";
  try {
    return format(new Date(dateString), "MMM d, yyyy");
  } catch {
    return "N/A";
  }
}

/**
 * Pluralize a number.
 *
 * @param count - The number to pluralize
 * @param singular - The singular form of the word
 * @param plural - The plural form of the word, defaults to singular + "s"
 * @returns The pluralized word
 */
export function pluralize(
  count: number,
  singular: string,
  plural?: string
): string {
  if (plural === undefined) plural = singular + "s";
  if (count === 1) return singular;
  return plural;
}
