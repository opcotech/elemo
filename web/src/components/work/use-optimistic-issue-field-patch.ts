import { useMutation, useQueryClient } from "@tanstack/react-query";

import {
  applyIssuePatchFields,
  beginOptimisticIssuePatch,
  commitIssueCaches,
  rollbackIssueCaches,
} from "./issue-cache-patch";
import { enqueueIssueUpdate } from "./issue-update-queue";

import { v1IssueUpdateMutation } from "@/lib/api/mutation-options";
import type { Issue, IssuePatch, PartialIssue } from "@/lib/api/types";
import { showErrorToast, showSuccessToast } from "@/lib/toast";
import type { WorkItem } from "@/lib/work/model";

function listFieldsFromUpdated(
  updated: Issue,
  patch: IssuePatch
): Partial<PartialIssue> {
  const fields: Partial<PartialIssue> = {};
  if ("status" in patch) {
    fields.status = updated.status;
  }
  if ("priority" in patch) {
    fields.priority = updated.priority;
  }
  if ("start_date" in patch) {
    fields.start_date = updated.start_date;
  }
  if ("due_date" in patch) {
    fields.due_date = updated.due_date;
  }
  return fields;
}

export function useOptimisticIssueFieldPatch(projectId: string | undefined) {
  const queryClient = useQueryClient();
  const mutation = useMutation(v1IssueUpdateMutation());

  const patchIssueFields = async ({
    item,
    patch,
    successDescription,
  }: {
    item: WorkItem;
    patch: IssuePatch;
    successDescription: string;
  }) => {
    if (item.dataSource !== "api") {
      return;
    }

    const resolvedProjectId = item.projectId ?? projectId;
    if (!resolvedProjectId) {
      const error = new Error("Issue is missing a project");
      showErrorToast("Failed to update issue", error.message);
      throw error;
    }

    await enqueueIssueUpdate(item.id, async () => {
      const snapshot = await beginOptimisticIssuePatch(
        queryClient,
        {
          issueId: item.id,
          projectId: resolvedProjectId,
          namespaceId: item.namespaceId,
          issueKey: item.key,
        },
        (issue) => applyIssuePatchFields(issue, patch)
      );

      try {
        const updated = await mutation.mutateAsync({
          path: { id: item.id },
          body: patch,
        });

        commitIssueCaches(
          queryClient,
          snapshot,
          updated,
          listFieldsFromUpdated(updated, patch)
        );

        showSuccessToast("Issue updated", successDescription);
      } catch (error) {
        rollbackIssueCaches(queryClient, snapshot, {
          invalidateProjectLists: true,
        });

        showErrorToast(
          "Failed to update issue",
          error instanceof Error ? error.message : "Unknown error occurred"
        );
        throw error;
      }
    });
  };

  return {
    patchIssueFields,
    isPending: mutation.isPending,
  };
}
