import {
  mockActivity,
  mockAttentionSignals,
  mockDocumentBodies,
  mockGlobalSearchEntries,
  mockPeople,
  mockRelations,
  mockSavedViews,
  mockTimeline,
  mockWorkItems,
} from "./fixtures";
import type {
  ActivityEntry,
  ActivityQuery,
  AttentionQuery,
  AttentionSignal,
  DocumentBody,
  GlobalSearchEntry,
  GlobalSearchQuery,
  PeopleQuery,
  Person,
  Relation,
  RelationQuery,
  SavedView,
  SavedViewQuery,
  Scope,
  SearchResultKind,
  TimelineEntry,
  TimelineQuery,
  WorkItem,
  WorkItemQuery,
  WorkPriority,
  WorkSort,
  WorkStatus,
} from "./types";

interface ScopedRecord {
  readonly namespaceId: string | null;
  readonly projectId: string | null;
  readonly personId?: string | null;
}

const priorityOrder: Record<WorkPriority, number> = {
  none: 0,
  low: 1,
  medium: 2,
  high: 3,
  urgent: 4,
};

const statusOrder: Record<WorkStatus, number> = {
  backlog: 0,
  planned: 1,
  "in-progress": 2,
  blocked: 3,
  done: 4,
  canceled: 5,
};

const severityOrder: Record<AttentionSignal["severity"], number> = {
  info: 0,
  warning: 1,
  critical: 2,
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

function workItemInScope(item: WorkItem, scope: Scope): boolean {
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

export function selectWorkItems(
  query: WorkItemQuery = {}
): readonly WorkItem[] {
  const scope = query.scope ?? { type: "global" };
  const sorts = query.sort?.length
    ? query.sort
    : ([{ field: "rank", direction: "asc" }] as const);

  return mockWorkItems
    .filter(
      (item) => workItemInScope(item, scope) && matchesWorkFilters(item, query)
    )
    .toSorted((left, right) => compareWorkItems(left, right, sorts));
}

export function getWorkItem(workItemId: string): WorkItem | undefined {
  return mockWorkItems.find((item) => item.id === workItemId);
}

function isViewAvailableInScope(
  view: SavedView,
  scope: Scope,
  includeGlobal: boolean
): boolean {
  if (view.scope.type === "global") {
    return includeGlobal || scope.type === "global";
  }
  if (scope.type === "global" || scope.type === "person") {
    return false;
  }
  if (scope.type === "namespace") {
    return (
      view.scope.type === "namespace" &&
      view.scope.namespaceId === scope.namespaceId
    );
  }

  return (
    (view.scope.type === "namespace" &&
      scope.namespaceId !== undefined &&
      view.scope.namespaceId === scope.namespaceId) ||
    (view.scope.type === "project" &&
      view.scope.projectId === scope.projectId &&
      (scope.namespaceId === undefined ||
        view.scope.namespaceId === undefined ||
        view.scope.namespaceId === scope.namespaceId))
  );
}

export function selectSavedViews(
  query: SavedViewQuery = {}
): readonly SavedView[] {
  const scope = query.scope ?? { type: "global" };
  const includeGlobal = query.includeGlobal ?? scope.type !== "global";

  return mockSavedViews
    .filter(
      (view) =>
        isViewAvailableInScope(view, scope, includeGlobal) &&
        (!query.ownerId || view.ownerId === query.ownerId) &&
        (!query.favoritesOnly || view.isFavorite)
    )
    .toSorted(
      (left, right) =>
        Number(right.isFavorite) - Number(left.isFavorite) ||
        left.name.localeCompare(right.name)
    );
}

export function getDocumentBody(documentId: string): DocumentBody | undefined {
  return mockDocumentBodies.find(
    (document) => document.documentId === documentId
  );
}

export function selectDocumentBodies(
  scope: Scope = { type: "global" }
): readonly DocumentBody[] {
  if (scope.type === "person") {
    return [];
  }

  return mockDocumentBodies.filter((document) =>
    isInScope(
      {
        namespaceId: document.namespaceId,
        projectId: document.projectId,
      },
      scope
    )
  );
}

export function selectRelations(query: RelationQuery): readonly Relation[] {
  const direction = query.direction ?? "either";

  return mockRelations
    .filter((relation) => {
      const incoming =
        relation.to.id === query.entity.id &&
        relation.to.type === query.entity.type;
      const outgoing =
        relation.from.id === query.entity.id &&
        relation.from.type === query.entity.type;
      const matchesDirection =
        direction === "incoming"
          ? incoming
          : direction === "outgoing"
            ? outgoing
            : incoming || outgoing;

      return (
        matchesDirection && includesWhenPresent(query.kinds, relation.kind)
      );
    })
    .toSorted(
      (left, right) =>
        right.createdAt.localeCompare(left.createdAt) ||
        left.id.localeCompare(right.id)
    );
}

export function selectActivity(
  query: ActivityQuery = {}
): readonly ActivityEntry[] {
  const entries = mockActivity
    .filter(
      (entry) =>
        (!query.entity ||
          (entry.entity.id === query.entity.id &&
            entry.entity.type === query.entity.type)) &&
        includesWhenPresent(query.actions, entry.action) &&
        includesWhenPresent(query.actorIds, entry.actorId)
    )
    .toSorted(
      (left, right) =>
        right.occurredAt.localeCompare(left.occurredAt) ||
        left.id.localeCompare(right.id)
    );

  return query.limit === undefined ? entries : entries.slice(0, query.limit);
}

export function selectTimeline(
  query: TimelineQuery = {}
): readonly TimelineEntry[] {
  const scope = query.scope ?? { type: "global" };

  return mockTimeline
    .filter((entry) => {
      const inPersonScope =
        scope.type !== "person" ||
        getWorkItem(entry.workItemId)?.assigneeId === scope.personId;
      return (
        inPersonScope &&
        (scope.type === "person" ||
          isInScope(
            {
              namespaceId: entry.namespaceId,
              projectId: entry.projectId,
            },
            scope
          )) &&
        (!query.from || entry.endAt >= query.from) &&
        (!query.to || entry.startAt <= query.to) &&
        includesWhenPresent(query.kinds, entry.kind)
      );
    })
    .toSorted(
      (left, right) =>
        left.startAt.localeCompare(right.startAt) ||
        left.id.localeCompare(right.id)
    );
}

export function selectAttentionSignals(
  query: AttentionQuery = {}
): readonly AttentionSignal[] {
  const scope = query.scope ?? { type: "global" };

  return mockAttentionSignals
    .filter(
      (signal) =>
        isInScope(
          {
            namespaceId: signal.namespaceId,
            projectId: signal.projectId,
            personId: signal.personId,
          },
          scope
        ) &&
        (!query.personId || signal.personId === query.personId) &&
        includesWhenPresent(query.reasons, signal.reason) &&
        includesWhenPresent(query.severities, signal.severity) &&
        (query.includeAcknowledged || signal.acknowledgedAt === null)
    )
    .toSorted(
      (left, right) =>
        severityOrder[right.severity] - severityOrder[left.severity] ||
        right.createdAt.localeCompare(left.createdAt) ||
        left.id.localeCompare(right.id)
    );
}

export function selectPeople(query: PeopleQuery = {}): readonly Person[] {
  const text = normalize(query.text ?? "");

  return mockPeople
    .filter((person) => {
      const searchableText = normalize(
        [
          person.displayName,
          person.handle,
          person.email,
          person.title,
          person.team,
        ].join(" ")
      );
      return (
        (!text || searchableText.includes(text)) &&
        includesWhenPresent(query.ids, person.id) &&
        includesWhenPresent(query.teams, person.team)
      );
    })
    .toSorted((left, right) =>
      left.displayName.localeCompare(right.displayName)
    );
}

export function getPerson(personId: string): Person | undefined {
  return mockPeople.find((person) => person.id === personId);
}

function searchEntryInScope(entry: GlobalSearchEntry, scope: Scope): boolean {
  if (scope.type === "global") {
    return true;
  }
  return isInScope(
    {
      namespaceId: entry.namespaceId,
      projectId: entry.projectId,
      personId: entry.personId,
    },
    scope
  );
}

function searchScore(entry: GlobalSearchEntry, text: string): number {
  const title = normalize(entry.title);
  const subtitle = normalize(entry.subtitle);
  const keywords = entry.keywords.map(normalize);

  if (title.startsWith(text)) {
    return 100;
  }
  if (title.includes(text)) {
    return 70;
  }
  if (keywords.some((keyword) => keyword === text)) {
    return 60;
  }
  if (keywords.some((keyword) => keyword.includes(text))) {
    return 40;
  }
  if (subtitle.includes(text)) {
    return 20;
  }
  return 0;
}

export function searchGlobalFixtures(
  query: GlobalSearchQuery
): readonly GlobalSearchEntry[] {
  const text = normalize(query.text);
  if (!text || query.limit === 0) {
    return [];
  }

  const scope = query.scope ?? { type: "global" };
  const kinds: readonly SearchResultKind[] | undefined = query.kinds;
  const terms = text.split(/\s+/);

  const matches = mockGlobalSearchEntries
    .filter((entry) => {
      const haystack = normalize(
        [entry.title, entry.subtitle, ...entry.keywords].join(" ")
      );
      return (
        searchEntryInScope(entry, scope) &&
        includesWhenPresent(kinds, entry.kind) &&
        terms.every((term) => haystack.includes(term))
      );
    })
    .map((entry) => ({ entry, score: searchScore(entry, text) }))
    .toSorted(
      (left, right) =>
        right.score - left.score ||
        left.entry.title.localeCompare(right.entry.title)
    )
    .map(({ entry }) => entry);

  return query.limit === undefined ? matches : matches.slice(0, query.limit);
}
