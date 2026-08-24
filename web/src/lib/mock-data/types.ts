import type {
  Scope,
  WorkFilters,
  WorkLayout,
  WorkSort,
} from "@/lib/work/model";

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

export interface MockRecord {
  readonly dataSource: "mock";
}

export type EntityType = "work-item" | "document" | "person";

export interface EntityRef {
  readonly id: string;
  readonly type: EntityType;
  readonly title: string;
}

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
