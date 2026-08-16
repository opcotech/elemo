import type { Label } from "@/lib/api/types";

export interface WorkLabel {
  readonly id: string;
  readonly name: string;
}

export function resolveWorkLabels(
  labelIds: readonly string[],
  labels?: readonly Pick<Label, "id" | "name">[] | null
): WorkLabel[] {
  const byId = new Map(
    (labels ?? []).map((label) => [label.id, label.name] as const)
  );
  const byName = new Map(
    (labels ?? []).map((label) => [label.name, label.name] as const)
  );

  return labelIds.map((id) => ({
    id,
    name: byId.get(id) ?? byName.get(id) ?? id,
  }));
}

/** Selected issue labels first (names already on the issue), then catalog extras. */
export function mergeWorkLabels(
  selected: readonly Pick<Label, "id" | "name">[],
  catalog?: readonly Pick<Label, "id" | "name">[] | null
): WorkLabel[] {
  const labels: WorkLabel[] = [];
  const seen = new Set<string>();
  const catalogById = new Map(
    (catalog ?? []).map((label) => [label.id, label.name] as const)
  );

  for (const label of selected) {
    if (seen.has(label.id)) {
      continue;
    }
    seen.add(label.id);
    const catalogName = catalogById.get(label.id);
    const name =
      catalogName && (!label.name || label.name === label.id)
        ? catalogName
        : label.name;
    labels.push({ id: label.id, name });
  }

  for (const label of catalog ?? []) {
    if (seen.has(label.id)) {
      continue;
    }
    seen.add(label.id);
    labels.push({ id: label.id, name: label.name });
  }

  return labels;
}

export function labelsFromIds(
  ids: readonly string[],
  ...sources: (readonly Pick<Label, "id" | "name">[] | null | undefined)[]
): WorkLabel[] {
  const byId = new Map<string, string>();

  for (const source of sources) {
    for (const label of source ?? []) {
      const current = byId.get(label.id);
      if (!current || current === label.id) {
        byId.set(label.id, label.name);
      }
    }
  }

  return ids.map((id) => ({ id, name: byId.get(id) ?? id }));
}
