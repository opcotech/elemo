import type { IssueKind, IssuePriority, IssueResolution } from "@/lib/client";

export const API_BACKED_DOMAINS = [
  "organizations",
  "namespaces",
  "projects",
  "issues",
  "todos",
  "notifications",
  "memberships",
  "administration",
] as const;

export const MOCK_ONLY_DOMAINS = [
  "workItems",
  "savedViews",
  "documentBodies",
  "relations",
  "activity",
  "timeline",
  "attentionSignals",
  "people",
  "globalSearch",
] as const;

export type ApiBackedDomain = (typeof API_BACKED_DOMAINS)[number];
export type MockOnlyDomain = (typeof MOCK_ONLY_DOMAINS)[number];

export type DataSource = "mock" | "api";

export interface MockRecord {
  readonly dataSource: "mock";
}

export type EntityType = "work-item" | "document" | "person";

export interface EntityRef {
  readonly id: string;
  readonly type: EntityType;
  readonly title: string;
}

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
  /** First assignee; kept for filters/grouping. */
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
  /**
   * ISO date-time used to evaluate due-date filters. Callers should pass their
   * current time; tests can pass a fixed value.
   */
  readonly now?: string;
}

export type WorkLayout = "board" | "list" | "table" | "timeline";

export interface SavedView extends MockRecord {
  readonly id: string;
  readonly name: string;
  readonly description: string;
  readonly ownerId: string;
  readonly scope: Scope;
  readonly layout: WorkLayout;
  readonly filters: WorkFilters;
  readonly sort: readonly WorkSort[];
  readonly groupBy: "status" | "priority" | "assignee" | "none";
  readonly isFavorite: boolean;
}

export interface SavedViewQuery {
  readonly scope?: Scope;
  readonly ownerId?: string;
  readonly favoritesOnly?: boolean;
  readonly includeGlobal?: boolean;
}

export type DocumentBlock =
  | {
      readonly id: string;
      readonly type: "heading";
      readonly level: 1 | 2 | 3;
      readonly text: string;
    }
  | {
      readonly id: string;
      readonly type: "paragraph";
      readonly text: string;
    }
  | {
      readonly id: string;
      readonly type: "checklist";
      readonly items: readonly {
        readonly text: string;
        readonly checked: boolean;
      }[];
    }
  | {
      readonly id: string;
      readonly type: "callout";
      readonly tone: "info" | "warning" | "success";
      readonly text: string;
    }
  | {
      readonly id: string;
      readonly type: "code";
      readonly language: string;
      readonly code: string;
    };

export interface DocumentBody extends MockRecord {
  readonly documentId: string;
  readonly title: string;
  readonly namespaceId: string;
  readonly projectId: string | null;
  readonly excerpt: string;
  readonly blocks: readonly DocumentBlock[];
  readonly updatedAt: string;
}

export type RelationKind =
  | "blocks"
  | "depends-on"
  | "implements"
  | "documents"
  | "references"
  | "related-to"
  | "owned-by";

export interface Relation extends MockRecord {
  readonly id: string;
  readonly kind: RelationKind;
  readonly from: EntityRef;
  readonly to: EntityRef;
  readonly createdAt: string;
  readonly createdBy: string;
}

export type RelationDirection = "incoming" | "outgoing" | "either";

export interface RelationQuery {
  readonly entity: Pick<EntityRef, "id" | "type">;
  readonly direction?: RelationDirection;
  readonly kinds?: readonly RelationKind[];
}

export type ActivityAction =
  | "created"
  | "updated"
  | "commented"
  | "assigned"
  | "status-changed"
  | "linked";

export interface ActivityEntry extends MockRecord {
  readonly id: string;
  readonly entity: EntityRef;
  readonly actorId: string;
  readonly action: ActivityAction;
  readonly detail: string;
  readonly occurredAt: string;
}

export interface ActivityQuery {
  readonly entity?: Pick<EntityRef, "id" | "type">;
  readonly actions?: readonly ActivityAction[];
  readonly actorIds?: readonly string[];
  readonly limit?: number;
}

export interface TimelineEntry extends MockRecord {
  readonly id: string;
  readonly workItemId: string;
  readonly namespaceId: string;
  readonly projectId: string | null;
  readonly title: string;
  readonly startAt: string;
  readonly endAt: string;
  readonly kind: "work" | "milestone";
}

export interface TimelineQuery {
  readonly scope?: Scope;
  readonly from?: string;
  readonly to?: string;
  readonly kinds?: readonly TimelineEntry["kind"][];
}

export type AttentionReason =
  "overdue" | "blocked" | "mentioned" | "unread-activity" | "due-soon";

export interface AttentionSignal extends MockRecord {
  readonly id: string;
  readonly workItemId: string;
  readonly namespaceId: string;
  readonly projectId: string | null;
  readonly personId: string;
  readonly reason: AttentionReason;
  readonly severity: "critical" | "warning" | "info";
  readonly summary: string;
  readonly createdAt: string;
  readonly acknowledgedAt: string | null;
}

export interface AttentionQuery {
  readonly scope?: Scope;
  readonly personId?: string;
  readonly reasons?: readonly AttentionReason[];
  readonly severities?: readonly AttentionSignal["severity"][];
  readonly includeAcknowledged?: boolean;
}

export interface Person extends MockRecord {
  readonly id: string;
  readonly displayName: string;
  readonly handle: string;
  readonly email: string;
  readonly avatarUrl: string | null;
  readonly title: string;
  readonly team: string;
}

export interface PeopleQuery {
  readonly text?: string;
  readonly ids?: readonly string[];
  readonly teams?: readonly string[];
}

export type SearchResultKind =
  "work-item" | "document" | "saved-view" | "person";

export interface GlobalSearchEntry extends MockRecord {
  readonly id: string;
  readonly kind: SearchResultKind;
  readonly title: string;
  readonly subtitle: string;
  readonly keywords: readonly string[];
  readonly namespaceId: string | null;
  readonly projectId: string | null;
  readonly personId: string | null;
  readonly href: string;
}

export interface GlobalSearchQuery {
  readonly text: string;
  readonly scope?: Scope;
  readonly kinds?: readonly SearchResultKind[];
  readonly limit?: number;
}
