import type { WorkItem } from "@/lib/mock-data";

export type WorkFieldOverride = Partial<
  Pick<WorkItem, "status" | "priority" | "startDate" | "dueDate">
>;

/** Drop only the failed fields so a later optimistic patch on the same issue stays. */
export function rollbackWorkFieldOverride(
  overrides: ReadonlyMap<string, WorkFieldOverride>,
  issueId: string,
  fields: readonly (keyof WorkFieldOverride)[]
): ReadonlyMap<string, WorkFieldOverride> {
  const current = overrides.get(issueId);
  if (!current) {
    return overrides;
  }

  let changed = false;
  const nextOverride: WorkFieldOverride = { ...current };
  for (const field of fields) {
    if (Object.hasOwn(nextOverride, field)) {
      delete nextOverride[field];
      changed = true;
    }
  }

  if (!changed) {
    return overrides;
  }

  const next = new Map(overrides);
  if (Object.keys(nextOverride).length === 0) {
    next.delete(issueId);
  } else {
    next.set(issueId, nextOverride);
  }
  return next;
}
