import type { QueryClient } from "@tanstack/react-query";
import { notFound } from "@tanstack/react-router";

import { isNotFoundOrForbidden } from "@/lib/api/errors";
import { v1NamespacesIssuesKeyGetOptions } from "@/lib/api/query-options";
import type { Issue } from "@/lib/api/types";
import { issueToWorkItem } from "@/lib/work/issue-adapter";
import type { WorkItem } from "@/lib/work/model";

export interface WorkItemLoaderData {
  readonly source: "api";
  readonly item: WorkItem;
  readonly issue: Issue;
  readonly namespaceId: string;
  readonly issueKey: string;
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
    if (isNotFoundOrForbidden(error)) {
      throw notFound();
    }
    throw error;
  }
}
