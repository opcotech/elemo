import { format } from "date-fns";

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
