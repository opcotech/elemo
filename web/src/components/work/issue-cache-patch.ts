import type {
  InfiniteData,
  Query,
  QueryClient,
  QueryKey,
} from "@tanstack/react-query";

import {
  v1IssueGetOptions,
  v1NamespacesIssuesKeyGetOptions,
} from "@/lib/api/query-options";
import { namespaceRefPath } from "@/lib/api/refs";
import type {
  Issue,
  IssuePatch,
  PartialIssue,
  PartialIssuePage,
} from "@/lib/api/types";
import { rollbackOptimisticQueryData } from "@/lib/mutation-workflow";

export type IssueCacheTarget = {
  issueId: string;
  projectId?: string | null;
  organizationId?: string | null;
  namespaceId?: string | null;
  issueKey?: string | null;
};

export type IssueCacheSnapshot = {
  target: IssueCacheTarget;
  byIdQueryKey: QueryKey;
  byKeyQueryKey: QueryKey | undefined;
  previousById: Issue | undefined;
  previousByKey: Issue | undefined;
  previousProjectLists: {
    queryKey: QueryKey;
    data: unknown;
  }[];
};

export function applyIssuePatchFields<T extends object>(
  issue: T,
  patch: IssuePatch
): T {
  return {
    ...issue,
    ...patch,
    updated_at: new Date().toISOString(),
  } as T;
}

export function isProjectsIssuesQueryForProject(
  query: Query,
  projectId: string
): boolean {
  const key = query.queryKey[0];
  if (!key || typeof key !== "object") {
    return false;
  }

  const record = key as {
    _id?: string;
    path?: { id?: string; projectId?: string };
  };

  return (
    record._id === "v1ProjectsIssuesGet" && record.path?.projectId === projectId
  );
}

export function mergePartialIssueFields(
  issue: PartialIssue,
  fields: Partial<PartialIssue>
): PartialIssue {
  return {
    ...issue,
    ...fields,
    assignees: fields.assignees ?? issue.assignees,
    reviewers: fields.reviewers ?? issue.reviewers,
    labels: fields.labels ?? issue.labels,
  };
}

function isPartialIssuePage(data: unknown): data is PartialIssuePage {
  return (
    typeof data === "object" &&
    data !== null &&
    Array.isArray((data as PartialIssuePage).items)
  );
}

function isInfinitePartialIssuePage(
  data: unknown
): data is InfiniteData<PartialIssuePage> {
  if (
    typeof data !== "object" ||
    data === null ||
    !Array.isArray((data as { pages?: unknown[] }).pages)
  ) {
    return false;
  }
  return (data as { pages: unknown[] }).pages.every(isPartialIssuePage);
}

export function patchProjectIssuesCaches(
  queryClient: QueryClient,
  projectId: string,
  issueId: string,
  updater: (issue: PartialIssue) => PartialIssue
): { queryKey: QueryKey; data: unknown }[] {
  const previousEntries: {
    queryKey: QueryKey;
    data: unknown;
  }[] = [];

  for (const query of queryClient.getQueryCache().getAll()) {
    if (!isProjectsIssuesQueryForProject(query, projectId)) {
      continue;
    }

    const queryKey = query.queryKey;
    const data = queryClient.getQueryData(queryKey);
    previousEntries.push({ queryKey, data });

    if (isPartialIssuePage(data)) {
      queryClient.setQueryData<PartialIssuePage>(queryKey, {
        ...data,
        items: data.items.map((issue) =>
          issue.id === issueId ? updater(issue) : issue
        ),
      });
      continue;
    }

    if (isInfinitePartialIssuePage(data)) {
      queryClient.setQueryData<InfiniteData<PartialIssuePage>>(queryKey, {
        ...data,
        pages: data.pages.map((page) => ({
          ...page,
          items: page.items.map((issue) =>
            issue.id === issueId ? updater(issue) : issue
          ),
        })),
      });
    }
  }

  return previousEntries;
}

function issueDetailQueryKeys(target: IssueCacheTarget): {
  byIdQueryKey: QueryKey;
  byKeyQueryKey: QueryKey | undefined;
} {
  const byIdQueryKey = v1IssueGetOptions({
    path: { id: target.issueId },
  }).queryKey;
  const byKeyQueryKey =
    target.organizationId && target.namespaceId && target.issueKey
      ? v1NamespacesIssuesKeyGetOptions({
          path: {
            ...namespaceRefPath(target.organizationId, target.namespaceId),
            key: target.issueKey,
          },
        }).queryKey
      : undefined;

  return { byIdQueryKey, byKeyQueryKey };
}

export async function cancelIssueCaches(
  queryClient: QueryClient,
  target: IssueCacheTarget,
  options?: { cancelDetailQueries?: boolean }
): Promise<void> {
  const tasks: Promise<unknown>[] = [];
  const { byIdQueryKey, byKeyQueryKey } = issueDetailQueryKeys(target);

  if (options?.cancelDetailQueries) {
    tasks.push(queryClient.cancelQueries({ queryKey: byIdQueryKey }));
    if (byKeyQueryKey) {
      tasks.push(queryClient.cancelQueries({ queryKey: byKeyQueryKey }));
    }
  }

  if (target.projectId) {
    const projectId = target.projectId;
    tasks.push(
      queryClient.cancelQueries({
        predicate: (query) => isProjectsIssuesQueryForProject(query, projectId),
      })
    );
  }

  await Promise.all(tasks);
}

export function snapshotAndPatchIssueCaches(
  queryClient: QueryClient,
  target: IssueCacheTarget,
  updater: <T extends object>(issue: T) => T
): IssueCacheSnapshot {
  const { byIdQueryKey, byKeyQueryKey } = issueDetailQueryKeys(target);

  const previousById = queryClient.getQueryData<Issue>(byIdQueryKey);
  if (previousById) {
    queryClient.setQueryData<Issue>(byIdQueryKey, updater(previousById));
  }

  const previousByKey = byKeyQueryKey
    ? queryClient.getQueryData<Issue>(byKeyQueryKey)
    : undefined;
  if (byKeyQueryKey && previousByKey) {
    queryClient.setQueryData<Issue>(byKeyQueryKey, updater(previousByKey));
  }

  const previousProjectLists = target.projectId
    ? patchProjectIssuesCaches(
        queryClient,
        target.projectId,
        target.issueId,
        (issue) => updater(issue)
      )
    : [];

  return {
    target,
    byIdQueryKey,
    byKeyQueryKey,
    previousById,
    previousByKey,
    previousProjectLists,
  };
}

export async function beginOptimisticIssuePatch(
  queryClient: QueryClient,
  target: IssueCacheTarget,
  updater: <T extends object>(issue: T) => T,
  options?: { cancelDetailQueries?: boolean }
): Promise<IssueCacheSnapshot> {
  await cancelIssueCaches(queryClient, target, options);
  return snapshotAndPatchIssueCaches(queryClient, target, updater);
}

export function rollbackIssueCaches(
  queryClient: QueryClient,
  snapshot: IssueCacheSnapshot,
  options?: { invalidateProjectLists?: boolean }
): void {
  rollbackOptimisticQueryData(queryClient, snapshot.byIdQueryKey, {
    previous: snapshot.previousById,
  });
  if (snapshot.byKeyQueryKey) {
    rollbackOptimisticQueryData(queryClient, snapshot.byKeyQueryKey, {
      previous: snapshot.previousByKey,
    });
  }
  for (const entry of snapshot.previousProjectLists) {
    queryClient.setQueryData(entry.queryKey, entry.data);
  }

  const projectId = snapshot.target.projectId;
  if (options?.invalidateProjectLists && projectId) {
    void queryClient.invalidateQueries({
      predicate: (query) => isProjectsIssuesQueryForProject(query, projectId),
    });
  }
}

export function commitIssueCaches(
  queryClient: QueryClient,
  snapshot: IssueCacheSnapshot,
  updated: Issue,
  listFields: Partial<PartialIssue>
): void {
  queryClient.setQueryData<Issue>(snapshot.byIdQueryKey, updated);
  if (snapshot.byKeyQueryKey) {
    queryClient.setQueryData<Issue>(snapshot.byKeyQueryKey, updated);
  }
  if (snapshot.target.projectId) {
    patchProjectIssuesCaches(
      queryClient,
      snapshot.target.projectId,
      snapshot.target.issueId,
      (issue) => mergePartialIssueFields(issue, listFields)
    );
  }
}
