import { z } from "zod";

import type { PartialProject } from "@/lib/api/types";

const titleSchema = z.string().min(3).max(120);
const descriptionPlainTextSchema = z.string().min(3);

/** Validate an issue title for inline commit. */
export function parseIssueTitle(
  value: string
): { ok: true; title: string } | { ok: false; error: string } {
  const trimmed = value.trim();
  const result = titleSchema.safeParse(trimmed);
  if (!result.success) {
    return {
      ok: false,
      error: result.error.issues[0]?.message ?? "Invalid title",
    };
  }
  return { ok: true, title: result.data };
}

/**
 * Normalize description for PATCH using plain-text length rules.
 * Empty plain text → null. Otherwise require min 10 on plain text and
 * persist the Markdown payload (falls back to plain text).
 */
export function parseIssueDescription(
  plainText: string,
  markdown?: string
): { ok: true; description: string | null } | { ok: false; error: string } {
  const plain = plainText.trim();
  if (plain.length === 0) {
    return { ok: true, description: null };
  }

  const result = descriptionPlainTextSchema.safeParse(plain);
  if (!result.success) {
    return {
      ok: false,
      error: result.error.issues[0]?.message ?? "Invalid description",
    };
  }

  const payload = (markdown ?? plainText).trim();
  return { ok: true, description: payload.length > 0 ? payload : null };
}

/** Match the owning project by composite issue key prefix (e.g. LMO-101 → LMO). */
export function matchProjectByIssueKey(
  issueKey: string,
  projects: readonly PartialProject[]
): PartialProject | undefined {
  const matches = projects.filter(
    (project) =>
      issueKey === project.key || issueKey.startsWith(`${project.key}-`)
  );
  if (matches.length === 0) {
    return undefined;
  }
  return [...matches].sort((a, b) => b.key.length - a.key.length)[0];
}

/**
 * Normalize multi-select assignment IDs for PATCH.
 * Drops blanks, dedupes, and preserves first-seen order. Empty → [].
 */
export function normalizeAssignmentIds(
  ids: readonly string[] | null | undefined
): string[] {
  const seen = new Set<string>();
  const result: string[] = [];

  for (const id of ids ?? []) {
    const trimmed = id.trim();
    if (!trimmed || seen.has(trimmed)) {
      continue;
    }
    seen.add(trimmed);
    result.push(trimmed);
  }

  return result;
}

/** True when both assignment ID lists contain the same members (order-insensitive). */
export function assignmentIdsEqual(
  a: readonly string[],
  b: readonly string[]
): boolean {
  if (a.length !== b.length) {
    return false;
  }
  const setB = new Set(b);
  return a.every((id) => setB.has(id));
}
