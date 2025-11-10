import { clsx } from "clsx";
import type { ClassValue } from "clsx";
import { format } from "date-fns";
import { twMerge } from "tailwind-merge";

export const SYSTEM_NIL_ID = "00000000000000000000";

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

/**
 * Get initials from a first and last name.
 *
 * @param firstName - The first name
 * @param lastName - The last name
 * @returns The initials in uppercase (e.g., "JD" for "John Doe")
 */
export function getInitials(firstName: string, lastName: string): string {
  return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase();
}

/**
 * Formats a resource ID to a human-readable format.
 *
 * @param resourceId - The resource ID to format
 * @returns The formatted resource ID
 */
export function formatResourceId(resourceId: string): string {
  if (resourceId === SYSTEM_NIL_ID) {
    return "System";
  }
  return resourceId;
}

/**
 * Returns the default value if the value is empty, otherwise returns the value.
 *
 * @param value - The value to check
 * @param defaultValue - The default value to return if the value is empty
 * @returns The default value if the value is empty, otherwise returns the value
 */
export function getDefaultValue<T extends string | undefined | null>(
  value: T,
  defaultValue: string = ""
): string {
  return value ?? defaultValue;
}

/**
 * Extracts the resource ID from a target string.
 *
 * @param target - The target string to extract the resource ID from
 * @returns The extracted resource ID
 */
export function extractResourceId(target: string): string {
  return target.split(":")[1] || target;
}

/**
 * Checks if a value is empty.
 *
 * @param value - The value to check
 * @returns True if the value is empty, false otherwise
 */
export function isEmpty<
  T extends string | undefined | null | any[] | Record<string, any>,
>(value: T): boolean {
  return (
    value === null ||
    value === undefined ||
    (typeof value === "string" && value.trim() === "") ||
    (Array.isArray(value) && value.length === 0) ||
    (typeof value === "object" && Object.keys(value).length === 0)
  );
}
