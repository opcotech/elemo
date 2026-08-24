import type { QueryKey } from "@tanstack/react-query";

import { collectListedPage, collectedQueryKey } from "@/lib/api/cursor-pages";
import {
  v1IssueGetOptions,
  v1IssueRelationsGetOptions,
  v1NamespacesIssuesGetOptions,
  v1NamespacesIssuesKeyGetOptions,
} from "@/lib/api/query-options";
import { v1NamespacesIssuesGet } from "@/lib/api/sdk";
import type {
  IssueRelation,
  IssueRelationDirection,
  IssueRelationKind,
  PartialIssue,
} from "@/lib/api/types";

export const ISSUE_RELATIONS_PAGE_SIZE = 100;
export const ISSUE_RELATIONS_PREVIEW_PAGE_SIZE = 4;

export const ISSUE_RELATION_KIND_SUBTASK_OF = "subtask of" as const;

export const editableIssueRelationKinds = [
  "blocked by",
  "blocks",
  "duplicated by",
  "duplicates",
  "related to",
] as const satisfies readonly IssueRelationKind[];

export type EditableIssueRelationKind =
  (typeof editableIssueRelationKinds)[number];

const inverseIssueRelationDisplayKind: Record<IssueRelationKind, string> = {
  "blocked by": "blocks",
  blocks: "blocked by",
  "depends on": "blocks",
  "duplicated by": "duplicates",
  duplicates: "duplicated by",
  "related to": "related to",
  "subtask of": ISSUE_RELATION_KIND_SUBTASK_OF,
};

export function isEditableIssueRelationKind(
  kind: string
): kind is EditableIssueRelationKind {
  return (editableIssueRelationKinds as readonly string[]).includes(kind);
}

export function issueRelationKindLabel(kind: string): string {
  if (kind.length === 0) {
    return kind;
  }
  return kind[0].toUpperCase() + kind.slice(1);
}

export function issueRelationDisplayKind(
  kind: IssueRelationKind,
  direction: IssueRelationDirection
): string {
  const canonical = kind === "depends on" ? "blocked by" : kind;
  if (direction === "outgoing") {
    return canonical;
  }
  return inverseIssueRelationDisplayKind[canonical];
}

export function issueRelationKindSelectValues(): string[] {
  return [...editableIssueRelationKinds];
}

export function relationKindPatch(
  currentDisplayKind: string,
  nextKind: string
): EditableIssueRelationKind | null {
  if (nextKind === currentDisplayKind) {
    return null;
  }
  if (!isEditableIssueRelationKind(nextKind)) {
    return null;
  }
  return nextKind;
}

export function visibleIssueRelations(
  relations: readonly IssueRelation[]
): IssueRelation[] {
  return relations.filter(
    (relation) => relation.kind !== ISSUE_RELATION_KIND_SUBTASK_OF
  );
}

export function relatedIssueIds(
  relations: readonly Pick<IssueRelation, "related">[]
): Set<string> {
  return new Set(relations.map((relation) => relation.related.id));
}

export function filterAvailableRelatedIssues<T extends { id: string }>(
  issues: readonly T[],
  currentIssueId: string,
  alreadyRelatedIds: ReadonlySet<string>
): T[] {
  return issues.filter(
    (issue) => issue.id !== currentIssueId && !alreadyRelatedIds.has(issue.id)
  );
}

export function relatedIssueCatalogQueryOptions(namespaceId: string) {
  const listOptions = v1NamespacesIssuesGetOptions({
    path: { id: namespaceId },
    query: { page_size: ISSUE_RELATIONS_PAGE_SIZE },
  });
  return {
    staleTime: listOptions.staleTime,
    gcTime: listOptions.gcTime,
    queryKey: collectedQueryKey(listOptions.queryKey),
    queryFn: async ({ signal }: { signal: AbortSignal }) =>
      collectListedPage(async (pageToken) => {
        const { data } = await v1NamespacesIssuesGet({
          path: { id: namespaceId },
          query: {
            page_size: ISSUE_RELATIONS_PAGE_SIZE,
            ...(pageToken ? { page_token: pageToken } : {}),
          },
          signal,
          throwOnError: true,
        });
        return data;
      }),
  };
}

export function relatedIssueWorkPath(
  related: Pick<PartialIssue, "key" | "namespace">,
  fallbackNamespaceId: string
): string {
  const namespaceId = related.namespace?.id ?? fallbackNamespaceId;
  return `/work/${namespaceId}/${related.key}`;
}

export function issueRelationInvalidationKeys({
  issueId,
  namespaceId,
  issueKey,
  related,
}: {
  issueId: string;
  namespaceId?: string;
  issueKey?: string;
  related?: Pick<PartialIssue, "id" | "key" | "namespace"> | null;
}): QueryKey[] {
  const keys: QueryKey[] = [
    v1IssueRelationsGetOptions({ path: { id: issueId } }).queryKey,
    v1IssueGetOptions({ path: { id: issueId } }).queryKey,
  ];

  if (namespaceId && issueKey) {
    keys.push(
      v1NamespacesIssuesKeyGetOptions({
        path: { id: namespaceId, key: issueKey },
      }).queryKey
    );
  }

  if (related) {
    keys.push(
      v1IssueRelationsGetOptions({ path: { id: related.id } }).queryKey,
      v1IssueGetOptions({ path: { id: related.id } }).queryKey
    );
    const relatedNamespaceId = related.namespace?.id;
    if (relatedNamespaceId && related.key) {
      keys.push(
        v1NamespacesIssuesKeyGetOptions({
          path: { id: relatedNamespaceId, key: related.key },
        }).queryKey
      );
    }
  }

  return keys;
}
