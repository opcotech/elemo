import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { QueryClient } from "@tanstack/react-query";

import {
  cancelIssueCaches,
  commitIssueCaches,
  isProjectsIssuesQueryForProject,
  rollbackIssueCaches,
  snapshotAndPatchIssueCaches,
} from "./issue-cache-patch";
import { enqueueIssueUpdate } from "./issue-update-queue";

import { v1IssueUpdateMutation } from "@/lib/api/mutation-options";
import {
  v1IssueGetOptions,
  v1LabelsGetOptions,
  v1NamespacesIssuesKeyGetOptions,
} from "@/lib/api/query-options";
import type {
  Issue,
  IssuePatch,
  Options,
  PartialLabel,
  PartialUser,
  V1IssueUpdateData,
} from "@/lib/api/types";
import type {
  LabelPage,
  OrganizationMemberPage,
  PartialIssue,
  PartialIssuePage,
} from "@/lib/client";
import { showErrorToast, showSuccessToast } from "@/lib/toast";
import { labelsFromIds } from "@/lib/work/resolve-work-labels";
import { partialUsersFromIds } from "@/lib/work/resolve-work-people";

interface UseIssueUpdateOptions {
  namespaceId: string;
  issueKey: string;
  issueId: string;
  projectId?: string | null;
}

function cachedMemberCatalog(queryClient: QueryClient): PartialUser[] {
  const users: PartialUser[] = [];
  for (const query of queryClient.getQueryCache().getAll()) {
    const key = query.queryKey[0];
    if (!key || typeof key !== "object") {
      continue;
    }
    if ((key as { _id?: string })._id !== "v1OrganizationMembersGet") {
      continue;
    }
    const data = queryClient.getQueryData<OrganizationMemberPage>(
      query.queryKey
    );
    for (const member of data?.items ?? []) {
      users.push({
        id: member.id,
        first_name: member.first_name,
        last_name: member.last_name,
        picture: member.picture,
      });
    }
  }
  return users;
}

function cachedLabelCatalog(queryClient: QueryClient): PartialLabel[] {
  const data = queryClient.getQueryData<LabelPage>(
    v1LabelsGetOptions().queryKey
  );
  return (data?.items ?? []).map((label) => ({
    id: label.id,
    name: label.name,
  }));
}

function applyIssuePatch<T extends object>(
  issue: T,
  patch: IssuePatch,
  labelCatalog: readonly PartialLabel[] = [],
  memberCatalog: readonly PartialUser[] = [],
  parentIssue?: PartialIssue | null
): T {
  const { parent, ...rest } = patch;
  const next = {
    ...issue,
    ...rest,
    updated_at: new Date().toISOString(),
    ...(parent !== undefined
      ? {
          parent:
            parent === null
              ? null
              : resolveParentIssue(parent, issue, parentIssue),
        }
      : {}),
  };

  if (patch.labels && hasPartialLabels(issue)) {
    (next as { labels: PartialLabel[] }).labels = labelsFromIds(
      patch.labels,
      issue.labels,
      labelCatalog
    );
  }

  if (patch.assignees && hasPartialUsers(issue, "assignees")) {
    (next as { assignees: PartialUser[] }).assignees = partialUsersFromIds(
      patch.assignees,
      issue.assignees,
      memberCatalog
    );
  }

  if (patch.reviewers && hasPartialUsers(issue, "reviewers")) {
    (next as { reviewers: PartialUser[] }).reviewers = partialUsersFromIds(
      patch.reviewers,
      issue.reviewers,
      memberCatalog
    );
  }

  return next as T;
}

function resolveParentIssue(
  parentId: string,
  issue: object,
  parentIssue?: PartialIssue | null
): PartialIssue {
  if (parentIssue?.id === parentId) {
    return parentIssue;
  }

  if (hasPartialIssueParent(issue) && issue.parent?.id === parentId) {
    return issue.parent;
  }

  return {
    id: parentId,
    key: parentId,
    numeric_id: 0,
    kind: "task",
    title: parentId,
    status: "open",
    priority: "normal",
    assignees: [],
    reviewers: [],
    labels: [],
  };
}

function hasPartialIssueParent(
  issue: object
): issue is { parent: PartialIssue | null | undefined } {
  if (!("parent" in issue)) {
    return false;
  }
  const value = issue.parent;
  return (
    value == null ||
    (typeof value === "object" && "id" in value && typeof value.id === "string")
  );
}

function cachedParentIssue(
  queryClient: QueryClient,
  projectId: string | null | undefined,
  parentId: string,
  currentParent?: PartialIssue | null
): PartialIssue {
  if (currentParent?.id === parentId) {
    return currentParent;
  }

  if (projectId) {
    for (const query of queryClient.getQueryCache().getAll()) {
      if (!isProjectsIssuesQueryForProject(query, projectId)) {
        continue;
      }
      const data = queryClient.getQueryData<PartialIssuePage>(query.queryKey);
      const found = data?.items.find((item) => item.id === parentId);
      if (found) {
        return found;
      }
    }
  }

  return resolveParentIssue(parentId, {}, currentParent);
}

function hasPartialUsers(
  issue: object,
  key: "assignees" | "reviewers"
): issue is { [K in typeof key]: readonly PartialUser[] } {
  const value = (issue as Record<string, unknown>)[key];
  return (
    Array.isArray(value) &&
    value.every(
      (user) =>
        user != null &&
        typeof user === "object" &&
        "id" in user &&
        "first_name" in user &&
        "last_name" in user
    )
  );
}

function hasPartialLabels(
  issue: object
): issue is { labels: readonly PartialLabel[] } {
  return (
    "labels" in issue &&
    Array.isArray(issue.labels) &&
    issue.labels.every(
      (label) =>
        label != null &&
        typeof label === "object" &&
        "id" in label &&
        "name" in label
    )
  );
}

export function useIssueUpdate({
  namespaceId,
  issueKey,
  issueId,
  projectId,
}: UseIssueUpdateOptions) {
  const queryClient = useQueryClient();
  const cacheTarget = { issueId, projectId, namespaceId, issueKey };

  const mutation = useMutation({
    ...v1IssueUpdateMutation(),
    onMutate: async (variables: Options<V1IssueUpdateData>) => {
      await cancelIssueCaches(queryClient, cacheTarget, {
        cancelDetailQueries: true,
      });

      const previousByKey = queryClient.getQueryData<Issue>(
        v1NamespacesIssuesKeyGetOptions({
          path: { id: namespaceId, key: issueKey },
        }).queryKey
      );
      const previousById = queryClient.getQueryData<Issue>(
        v1IssueGetOptions({
          path: { id: issueId },
        }).queryKey
      );
      const patch = variables.body ?? {};
      const labelCatalog = cachedLabelCatalog(queryClient);
      const memberCatalog = cachedMemberCatalog(queryClient);
      const parentIssue =
        typeof patch.parent === "string"
          ? cachedParentIssue(
              queryClient,
              projectId,
              patch.parent,
              previousByKey?.parent ?? previousById?.parent
            )
          : undefined;

      return snapshotAndPatchIssueCaches(queryClient, cacheTarget, (issue) =>
        applyIssuePatch(issue, patch, labelCatalog, memberCatalog, parentIssue)
      );
    },
    onError: (_error, _variables, context) => {
      if (context) {
        rollbackIssueCaches(queryClient, context);
      }
    },
    onSuccess: (data, _variables, context) => {
      if (context) {
        commitIssueCaches(queryClient, context, data, {
          kind: data.kind,
          title: data.title,
          description: data.description,
          status: data.status,
          priority: data.priority,
          parent: data.parent,
          assignees: data.assignees,
          reviewers: data.reviewers,
          labels: data.labels,
          due_date: data.due_date,
          start_date: data.start_date,
        });
      }
    },
  });

  const updateIssue = async (
    patch: IssuePatch,
    successDescription?: string
  ) => {
    try {
      await enqueueIssueUpdate(issueId, () =>
        mutation.mutateAsync({
          path: { id: issueId },
          body: patch,
        })
      );
      showSuccessToast(
        "Issue updated",
        successDescription ?? "Your changes were saved"
      );
    } catch (error) {
      showErrorToast(
        "Failed to update issue",
        error instanceof Error ? error.message : "Unknown error occurred"
      );
      throw error;
    }
  };

  return {
    updateIssue,
    isPending: mutation.isPending,
  };
}
