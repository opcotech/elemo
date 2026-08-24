import type {
  IssueKind,
  IssuePriority,
  IssueResolution,
} from "@/lib/api/types";

export type DataSource = "mock" | "api";

export type Scope =
  | { readonly type: "global" }
  | { readonly type: "namespace"; readonly namespaceId: string }
  | {
      readonly type: "project";
      readonly namespaceId?: string;
      readonly projectId: string;
    }
  | { readonly type: "person"; readonly personId: string };

export type WorkStatus =
  "backlog" | "in progress" | "in review" | "blocked" | "done" | "closed";

export type WorkPriority = IssuePriority;

export interface WorkPerson {
  readonly id: string;
  readonly name: string;
  readonly picture?: string | null;
}

export interface WorkNamespaceRef {
  readonly id: string;
  readonly name: string;
}

export interface WorkProjectRef {
  readonly id: string;
  readonly key: string;
  readonly name: string;
}

export interface WorkParentRef {
  readonly id: string;
  readonly key: string;
  readonly title: string;
  readonly namespaceId?: string;
}

export interface WorkItem {
  readonly dataSource: DataSource;
  readonly id: string;
  readonly key: string;
  readonly title: string;
  readonly summary: string;
  readonly namespaceId: string;
  readonly projectId: string | null;
  readonly namespace?: WorkNamespaceRef;
  readonly project?: WorkProjectRef;
  readonly kind?: IssueKind;
  readonly resolution?: IssueResolution;
  readonly parent?: WorkParentRef | null;
  readonly links?: readonly {
    readonly url: string;
    readonly label: string;
  }[];
  readonly status: WorkStatus;
  readonly priority: WorkPriority;
  readonly assigneeIds: readonly string[];
  readonly reviewerIds: readonly string[];
  readonly assignees?: readonly WorkPerson[];
  readonly reviewers?: readonly WorkPerson[];
  readonly assigneeId: string | null;
  readonly creatorId: string;
  readonly labelIds: readonly string[];
  readonly labels?: readonly { readonly id: string; readonly name: string }[];
  readonly rank: number;
  readonly dueDate: string | null;
  readonly startDate: string | null;
  readonly createdAt: string;
  readonly updatedAt: string;
}

export type WorkSortField =
  | "rank"
  | "title"
  | "priority"
  | "status"
  | "dueDate"
  | "createdAt"
  | "updatedAt";

export interface WorkSort {
  readonly field: WorkSortField;
  readonly direction: "asc" | "desc";
}

export type DueDateFilter = "overdue" | "today" | "upcoming" | "none";

export interface WorkFilters {
  readonly text?: string;
  readonly statuses?: readonly WorkStatus[];
  readonly priorities?: readonly WorkPriority[];
  readonly assigneeIds?: readonly string[];
  readonly labelIds?: readonly string[];
  readonly dueDate?: DueDateFilter;
}

export interface WorkItemQuery {
  readonly scope?: Scope;
  readonly filters?: WorkFilters;
  readonly sort?: readonly WorkSort[];
  readonly now?: string;
}

export type WorkLayout = "board" | "list" | "table" | "timeline";
