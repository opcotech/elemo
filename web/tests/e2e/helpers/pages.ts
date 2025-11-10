import type { Page } from "@playwright/test";

/**
 * Extract the ID from a URL.
 *
 * @param url - The URL to extract the ID from
 * @returns The extracted ID
 */
export function extractIdFromPath(
  page: Page,
  pattern: RegExp,
  nth: number = 0
): string {
  const path = new URL(page.url()).pathname;
  const idMatch = path.match(pattern);
  const group = nth + 1;

  if (!idMatch || !idMatch[group]) {
    throw new Error(
      `Could not extract ID from path: ${path} with pattern: ${pattern} and nth: ${group}`
    );
  }

  return idMatch[group];
}
