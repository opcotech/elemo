import type { IssuePriority, IssueStatus } from "@/lib/api/types";
import { mapWorkStatusToIssueStatus } from "@/lib/work/issue-adapter";
import type { WorkFilters, WorkSortField } from "@/lib/work/model";
import type { WorkRouteSearch } from "@/lib/work-route-search";

const sortableFields = new Set<WorkSortField>([
  "rank",
  "title",
  "priority",
  "status",
  "dueDate",
  "createdAt",
  "updatedAt",
]);

type IssueListOrderField =
  | "rank"
  | "numeric_id"
  | "title"
  | "priority"
  | "status"
  | "due_date"
  | "created_at"
  | "updated_at";

type IssueListOrder = `${IssueListOrderField}:${"asc" | "desc"}`;

function parseSort(sort: string): {
  field: WorkSortField;
  direction: "asc" | "desc";
} {
  const [fieldRaw, directionRaw] = sort.split(":");
  const field = sortableFields.has(fieldRaw as WorkSortField)
    ? (fieldRaw as WorkSortField)
    : "rank";
  return {
    field,
    direction: directionRaw === "desc" ? "desc" : "asc",
  };
}

function toIssueOrder(sort: WorkRouteSearch["sort"]): IssueListOrder {
  const { field, direction } = parseSort(sort);
  const apiField =
    field === "dueDate"
      ? "due_date"
      : field === "createdAt"
        ? "created_at"
        : field === "updatedAt"
          ? "updated_at"
          : field;
  return `${apiField}:${direction}` as IssueListOrder;
}

function toIssueStatuses(
  statuses: WorkFilters["statuses"]
): IssueStatus[] | undefined {
  if (!statuses?.length) {
    return undefined;
  }
  return statuses.map((status) => mapWorkStatusToIssueStatus(status));
}

function toIssuePriorities(
  priorities: WorkFilters["priorities"]
): IssuePriority[] | undefined {
  if (!priorities?.length) {
    return undefined;
  }
  return [...priorities];
}

export type IssueListApiQuery = {
  q?: string;
  status?: IssueStatus[];
  priority?: IssuePriority[];
  order: IssueListOrder;
};

export function buildIssueListApiQuery(
  search: WorkRouteSearch,
  activeViewFilters?: WorkFilters
): IssueListApiQuery {
  const text = (search.filter ?? activeViewFilters?.text)?.trim() ?? "";
  const statuses = toIssueStatuses(activeViewFilters?.statuses);
  const priorities = toIssuePriorities(activeViewFilters?.priorities);
  return {
    ...(text ? { q: text } : {}),
    ...(statuses ? { status: statuses } : {}),
    ...(priorities ? { priority: priorities } : {}),
    order: toIssueOrder(search.sort),
  };
}

export function issueListClientOnlyFilters(
  activeViewFilters?: WorkFilters
): WorkFilters {
  return {
    assigneeIds: activeViewFilters?.assigneeIds,
    labelIds: activeViewFilters?.labelIds,
    dueDate: activeViewFilters?.dueDate,
  };
}
