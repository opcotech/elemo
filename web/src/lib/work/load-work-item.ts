import type { QueryClient } from "@tanstack/react-query";
import { notFound } from "@tanstack/react-router";

import { isApiError, isNotFound, isPermissionDenied } from "@/lib/api/errors";
import { v1NamespacesIssuesKeyGetOptions } from "@/lib/api/query-options";
import type { Issue } from "@/lib/api/types";
import { getWorkItem } from "@/lib/mock-data";
import type { WorkItem } from "@/lib/mock-data";
import { issueToWorkItem } from "@/lib/work/issue-adapter";

export type WorkItemLoaderData =
  | {
      source: "mock";
      item: WorkItem;
    }
  | {
      source: "api";
      item: WorkItem;
      issue: Issue;
      namespaceId: string;
      issueKey: string;
    };

function isClientRequestError(error: unknown): boolean {
  if (isNotFound(error) || isPermissionDenied(error)) {
    return true;
  }
  if (isApiError(error)) {
    return error.status >= 400 && error.status < 500;
  }
  if (
    error &&
    typeof error === "object" &&
    "status" in error &&
    typeof error.status === "number"
  ) {
    return error.status >= 400 && error.status < 500;
  }
  return false;
}

export async function loadWorkItemPage(
  queryClient: QueryClient,
  namespaceId: string,
  issueKey: string
): Promise<WorkItemLoaderData> {
  try {
    const issue = await queryClient.fetchQuery(
      v1NamespacesIssuesKeyGetOptions({
        path: { id: namespaceId, key: issueKey },
      })
    );
    return {
      source: "api",
      issue,
      namespaceId,
      issueKey,
      item: issueToWorkItem(issue, { namespaceId }),
    };
  } catch (error) {
    if (isNotFound(error)) {
      const fixtureItem = getWorkItem(issueKey, namespaceId);
      if (fixtureItem) {
        return { source: "mock", item: fixtureItem };
      }
      throw notFound();
    }
    if (isClientRequestError(error)) {
      throw notFound();
    }
    throw error;
  }
}
