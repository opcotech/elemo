import type {
  Scope,
  WorkItem,
  WorkItemQuery,
  WorkPriority,
  WorkSort,
  WorkStatus,
} from "@/lib/work/model";

interface ScopedRecord {
  readonly namespaceId: string | null;
  readonly projectId: string | null;
  readonly personId?: string | null;
}

const priorityOrder: Record<WorkPriority, number> = {
  lowest: 0,
  low: 1,
  normal: 2,
  high: 3,
  highest: 4,
};

const statusOrder: Record<WorkStatus, number> = {
  backlog: 0,
  "in progress": 1,
  "in review": 2,
  blocked: 3,
  done: 4,
  closed: 5,
};

function normalize(value: string): string {
  return value.trim().toLocaleLowerCase();
}

function includesWhenPresent<T>(
  options: readonly T[] | undefined,
  value: T
): boolean {
  return !options?.length || options.includes(value);
}

export function isInScope(record: ScopedRecord, scope: Scope): boolean {
  switch (scope.type) {
    case "global":
      return true;
    case "namespace":
      return record.namespaceId === scope.namespaceId;
    case "project":
      return (
        record.projectId === scope.projectId &&
        (scope.namespaceId === undefined ||
          record.namespaceId === scope.namespaceId)
      );
    case "person":
      return record.personId === scope.personId;
  }
}

export function isWorkItemInScope(item: WorkItem, scope: Scope): boolean {
  return isInScope(
    {
      namespaceId: item.namespaceId,
      projectId: item.projectId,
      personId: item.assigneeId,
    },
    scope
  );
}

function matchesDueDate(
  item: WorkItem,
  dueDate: NonNullable<WorkItemQuery["filters"]>["dueDate"],
  now: string
): boolean {
  if (dueDate === undefined) {
    return true;
  }
  if (dueDate === "none") {
    return item.dueDate === null;
  }
  if (item.dueDate === null) {
    return false;
  }

  const dueDay = item.dueDate.slice(0, 10);
  const today = now.slice(0, 10);

  if (dueDate === "today") {
    return dueDay === today;
  }
  if (dueDate === "overdue") {
    return dueDay < today;
  }
  return dueDay > today;
}

function matchesWorkFilters(item: WorkItem, query: WorkItemQuery): boolean {
  const filters = query.filters;
  if (!filters) {
    return true;
  }

  const text = normalize(filters.text ?? "");
  const searchableText = normalize(
    [item.key, item.title, item.summary, ...item.labelIds].join(" ")
  );

  return (
    (!text || searchableText.includes(text)) &&
    includesWhenPresent(filters.statuses, item.status) &&
    includesWhenPresent(filters.priorities, item.priority) &&
    (!filters.assigneeIds?.length ||
      (item.assigneeId !== null &&
        filters.assigneeIds.includes(item.assigneeId))) &&
    (!filters.labelIds?.length ||
      filters.labelIds.every((labelId) => item.labelIds.includes(labelId))) &&
    matchesDueDate(item, filters.dueDate, query.now ?? new Date().toISOString())
  );
}

function compareNullable<T extends number | string>(
  left: T | null,
  right: T | null,
  direction: WorkSort["direction"]
): number {
  if (left === null && right === null) {
    return 0;
  }
  if (left === null) {
    return 1;
  }
  if (right === null) {
    return -1;
  }

  const comparison =
    typeof left === "number" && typeof right === "number"
      ? left - right
      : String(left).localeCompare(String(right));
  return direction === "asc" ? comparison : -comparison;
}

function compareWorkItems(
  left: WorkItem,
  right: WorkItem,
  sorts: readonly WorkSort[]
): number {
  for (const sort of sorts) {
    let comparison = 0;
    switch (sort.field) {
      case "priority":
        comparison = compareNullable(
          priorityOrder[left.priority],
          priorityOrder[right.priority],
          sort.direction
        );
        break;
      case "status":
        comparison = compareNullable(
          statusOrder[left.status],
          statusOrder[right.status],
          sort.direction
        );
        break;
      case "rank":
        comparison = compareNullable(left.rank, right.rank, sort.direction);
        break;
      case "dueDate":
        comparison = compareNullable(
          left.dueDate,
          right.dueDate,
          sort.direction
        );
        break;
      case "title":
      case "createdAt":
      case "updatedAt":
        comparison = compareNullable(
          left[sort.field],
          right[sort.field],
          sort.direction
        );
        break;
    }
    if (comparison !== 0) {
      return comparison;
    }
  }

  return left.id.localeCompare(right.id);
}

export function queryWorkItems(
  items: readonly WorkItem[],
  query: WorkItemQuery = {}
): readonly WorkItem[] {
  const scope = query.scope;
  const sorts = query.sort?.length
    ? query.sort
    : ([{ field: "rank", direction: "asc" }] as const);

  return items
    .filter(
      (item) =>
        (!scope || isWorkItemInScope(item, scope)) &&
        matchesWorkFilters(item, query)
    )
    .toSorted((left, right) => compareWorkItems(left, right, sorts));
}
