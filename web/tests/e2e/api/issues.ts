import { withErrorHandling } from "./error-handler";

import type { Client } from "@/lib/api/client";
import { projectIdPath } from "@/lib/api/refs";
import {
  v1IssueGet,
  v1IssueRelationsCreate,
  v1IssueUpdate,
  v1ProjectsIssuesCreate,
} from "@/lib/api/sdk";
import type {
  Issue,
  IssueCreate,
  IssueKind,
  IssuePatch,
  IssueRelation,
  IssueRelationCreate,
} from "@/lib/api/types";

/**
 * Create an issue via API, then fetch the full issue by ID.
 *
 * @param client - Authenticated API client
 * @param projectId - Project ID to create the issue in
 * @param issueData - Issue data (title is required; kind defaults to task)
 * @returns Created issue with ID and key
 */
export async function createIssue(
  client: Client,
  projectId: string,
  issueData: Partial<IssueCreate> & { title: string; kind?: IssueKind }
): Promise<Issue> {
  const issueCreateData: IssueCreate = {
    kind: issueData.kind ?? "task",
    title: issueData.title,
    description: issueData.description,
    parent: issueData.parent,
    status: issueData.status,
    priority: issueData.priority,
    resolution: issueData.resolution,
    links: issueData.links,
    due_date: issueData.due_date,
    start_date: issueData.start_date,
  };

  const response = await withErrorHandling(
    async () => {
      return await v1ProjectsIssuesCreate({
        client,
        path: projectIdPath(projectId),
        body: issueCreateData,
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/projects/${projectId}/issues`,
      method: "POST",
    }
  );

  const issueId = response.data.id || "";

  const issueResponse = await withErrorHandling(
    async () => {
      return await v1IssueGet({
        client,
        path: { id: issueId },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/issues/${issueId}`,
      method: "GET",
    }
  );

  return issueResponse.data;
}

/**
 * Update an issue via API.
 *
 * @param client - Authenticated API client
 * @param issueId - Issue ID to update
 * @param patch - Fields to change
 * @returns Updated issue
 */
export async function updateIssue(
  client: Client,
  issueId: string,
  patch: IssuePatch
): Promise<Issue> {
  const response = await withErrorHandling(
    async () => {
      return await v1IssueUpdate({
        client,
        path: { id: issueId },
        body: patch,
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/issues/${issueId}`,
      method: "PATCH",
    }
  );

  return response.data;
}

/**
 * Create an outgoing relation from one issue to another.
 *
 * @param client - Authenticated API client
 * @param issueId - Source issue ID
 * @param relationData - Related issue ID and kind
 * @returns Created relation
 */
export async function createIssueRelation(
  client: Client,
  issueId: string,
  relationData: IssueRelationCreate
): Promise<IssueRelation> {
  const response = await withErrorHandling(
    async () => {
      return await v1IssueRelationsCreate({
        client,
        path: { id: issueId },
        body: relationData,
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/issues/${issueId}/relations`,
      method: "POST",
    }
  );

  return response.data;
}
