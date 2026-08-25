import type { QueryClient } from "@tanstack/react-query";
import { notFound } from "@tanstack/react-router";

import { isNotFoundOrForbidden } from "@/lib/api/errors";
import { v1NamespacesIssuesKeyGetOptions } from "@/lib/api/query-options";
import { namespaceRefPath } from "@/lib/api/refs";
import type { Issue } from "@/lib/api/types";
import { loadAccessibleNamespace } from "@/lib/operational-route-data";
import {
  requireIssueKey,
  requireNamespaceSlug,
  requireOrganizationSlug,
} from "@/lib/route-identity";
import { issueToWorkItem } from "@/lib/work/issue-adapter";
import type { WorkItem } from "@/lib/work/model";

export interface WorkItemLoaderData {
  readonly source: "api";
  readonly item: WorkItem;
  readonly issue: Issue;
  readonly organizationSlug: string;
  readonly namespaceSlug: string;
  readonly organizationId: string;
  readonly namespaceId: string;
  readonly issueKey: string;
}

export async function loadWorkItemPage(
  queryClient: QueryClient,
  organizationSlug: string,
  namespaceSlug: string,
  issueKey: string
): Promise<WorkItemLoaderData> {
  requireOrganizationSlug(organizationSlug);
  requireNamespaceSlug(namespaceSlug);
  requireIssueKey(issueKey);

  try {
    const [issue, namespace] = await Promise.all([
      queryClient.fetchQuery(
        v1NamespacesIssuesKeyGetOptions({
          path: {
            ...namespaceRefPath(organizationSlug, namespaceSlug),
            key: issueKey,
          },
        })
      ),
      loadAccessibleNamespace(queryClient, organizationSlug, namespaceSlug),
    ]);

    if (issue.namespace?.id && issue.namespace.id !== namespace.id) {
      throw notFound();
    }

    return {
      source: "api",
      issue,
      organizationSlug,
      namespaceSlug,
      organizationId: namespace.organization.id,
      namespaceId: namespace.id,
      issueKey,
      item: issueToWorkItem(issue, {
        namespaceId: namespace.id,
        organizationId: namespace.organization.id,
        organizationSlug,
        namespaceSlug,
      }),
    };
  } catch (error) {
    if (isNotFoundOrForbidden(error)) {
      throw notFound();
    }
    throw error;
  }
}
