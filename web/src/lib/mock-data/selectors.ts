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
  SearchResultKind,
  TimelineEntry,
  TimelineQuery,
} from "./types";

import type { Scope, WorkItem, WorkItemQuery } from "@/lib/work/model";
import { isInScope, queryWorkItems } from "@/lib/work/query";

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

export function selectWorkItems(
  query: WorkItemQuery = {}
): readonly WorkItem[] {
  return queryWorkItems(mockWorkItems, query);
}

export function getWorkItem(
  workItemId: string,
  namespaceId?: string
): WorkItem | undefined {
  return mockWorkItems.find((item) => {
    if (namespaceId && item.namespaceId !== namespaceId) {
      return false;
    }
    return item.id === workItemId || item.key === workItemId;
  });
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
