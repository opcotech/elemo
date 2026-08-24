import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it } from "vitest";

import {
  applyIssuePatchFields,
  commitIssueCaches,
  isProjectsIssuesQueryForProject,
  mergePartialIssueFields,
  patchProjectIssuesCaches,
  rollbackIssueCaches,
  snapshotAndPatchIssueCaches,
} from "./issue-cache-patch";

import {
  v1IssueGetOptions,
  v1NamespacesIssuesKeyGetOptions,
  v1ProjectsIssuesGetOptions,
} from "@/lib/api/query-options";
import type { Issue, PartialIssue, PartialIssuePage } from "@/lib/api/types";

function partialIssue(overrides: Partial<PartialIssue> = {}): PartialIssue {
  return {
    id: "issue-1",
    key: "PLAT-1",
    numeric_id: 1,
    kind: "task",
    title: "Move me",
    status: "open",
    priority: "normal",
    assignees: [],
    reviewers: [],
    labels: [],
    created_at: "2026-01-01T00:00:00.000Z",
    updated_at: "2026-01-01T00:00:00.000Z",
    ...overrides,
  };
}

function issuePage(
  items: PartialIssue[],
  pageInfo: PartialIssuePage["page_info"] = { has_more: false }
): PartialIssuePage {
  return { items, page_info: pageInfo };
}

function detailIssue(overrides: Partial<Issue> = {}): Issue {
  return {
    id: "issue-1",
    key: "PLAT-1",
    numeric_id: 1,
    kind: "task",
    title: "Move me",
    description: "",
    status: "open",
    priority: "normal",
    resolution: "none",
    reported_by: {
      id: "user-1",
      first_name: "Test",
      last_name: "User",
      picture: null,
    },
    assignees: [],
    reviewers: [],
    labels: [],
    parent: null,
    links: [],
    comment_count: 0,
    attachment_count: 0,
    watcher_count: 0,
    relation_count: 0,
    due_date: null,
    start_date: null,
    created_at: "2026-01-01T00:00:00.000Z",
    updated_at: "2026-01-01T00:00:00.000Z",
    ...overrides,
  };
}

describe("project issues cache patching", () => {
  it("matches project issue list queries by generated key id and path", () => {
    const queryClient = new QueryClient();
    const matching = v1ProjectsIssuesGetOptions({
      path: { id: "project-1" },
      query: { page_size: 100 },
    });
    const otherProject = v1ProjectsIssuesGetOptions({
      path: { id: "project-2" },
      query: { page_size: 100 },
    });

    queryClient.setQueryData(matching.queryKey, issuePage([]));
    queryClient.setQueryData(otherProject.queryKey, issuePage([]));

    const queries = queryClient.getQueryCache().getAll();
    const matched = queries.filter((query) =>
      isProjectsIssuesQueryForProject(query, "project-1")
    );

    expect(matched).toHaveLength(1);
    expect(matched[0]?.queryKey).toEqual(matching.queryKey);
  });

  it("updates the matching issue inside a paginated project issues page", () => {
    const queryClient = new QueryClient();
    const options = v1ProjectsIssuesGetOptions({
      path: { id: "project-1" },
      query: { page_size: 100 },
    });
    const page = issuePage([
      partialIssue({ id: "issue-1", status: "open" }),
      partialIssue({
        id: "issue-2",
        key: "PLAT-2",
        numeric_id: 2,
        status: "open",
      }),
    ]);
    queryClient.setQueryData(options.queryKey, page);

    const previous = patchProjectIssuesCaches(
      queryClient,
      "project-1",
      "issue-1",
      (issue) => ({ ...issue, status: "done" })
    );

    expect(
      queryClient.getQueryData<PartialIssuePage>(options.queryKey)
    ).toEqual({
      items: [{ ...page.items[0], status: "done" }, page.items[1]],
      page_info: page.page_info,
    });
    expect(previous).toEqual([{ queryKey: options.queryKey, data: page }]);
  });

  it("leaves a paginated page unchanged when the issue is not in the page", () => {
    const queryClient = new QueryClient();
    const options = v1ProjectsIssuesGetOptions({
      path: { id: "project-1" },
      query: { page_size: 100 },
    });
    const page = issuePage([partialIssue({ id: "issue-2", status: "open" })]);
    queryClient.setQueryData(options.queryKey, page);

    patchProjectIssuesCaches(queryClient, "project-1", "issue-1", (issue) => ({
      ...issue,
      status: "done",
    }));

    expect(queryClient.getQueryData(options.queryKey)).toEqual(page);
  });

  it("merges status without replacing list item fields or page shape", () => {
    const queryClient = new QueryClient();
    const options = v1ProjectsIssuesGetOptions({
      path: { id: "project-1" },
      query: { page_size: 100 },
    });
    const existing = partialIssue({
      id: "issue-1",
      title: "Keep my title",
      status: "open",
      assignees: [{ id: "user-2", first_name: "Ada", last_name: "Lovelace" }],
      reviewers: [{ id: "user-4", first_name: "Grace", last_name: "Hopper" }],
      labels: [{ id: "label-1", name: "frontend" }],
    });
    const page = issuePage([existing]);
    queryClient.setQueryData(options.queryKey, page);

    patchProjectIssuesCaches(queryClient, "project-1", "issue-1", (issue) =>
      mergePartialIssueFields(issue, { status: "done", priority: "high" })
    );

    const next = queryClient.getQueryData<PartialIssuePage>(options.queryKey);
    expect(next).toEqual({
      items: [
        {
          ...existing,
          status: "done",
          priority: "high",
        },
      ],
      page_info: page.page_info,
    });
    expect(Array.isArray(next?.items)).toBe(true);
    expect(next?.items[0]?.assignees).toEqual([
      { id: "user-2", first_name: "Ada", last_name: "Lovelace" },
    ]);
    expect(next?.items[0]?.title).toBe("Keep my title");
  });
});

describe("optimistic issue cache snapshot", () => {
  it("patches detail and list caches then restores them on rollback", () => {
    const queryClient = new QueryClient();
    const listOptions = v1ProjectsIssuesGetOptions({
      path: { id: "project-1" },
      query: { page_size: 100 },
    });
    const byIdOptions = v1IssueGetOptions({ path: { id: "issue-1" } });
    const byKeyOptions = v1NamespacesIssuesKeyGetOptions({
      path: { id: "ns-1", key: "PLAT-1" },
    });
    const page = issuePage([partialIssue({ status: "open" })]);
    const original = detailIssue({ status: "open" });

    queryClient.setQueryData(listOptions.queryKey, page);
    queryClient.setQueryData(byIdOptions.queryKey, original);
    queryClient.setQueryData(byKeyOptions.queryKey, original);

    const snapshot = snapshotAndPatchIssueCaches(
      queryClient,
      {
        issueId: "issue-1",
        projectId: "project-1",
        namespaceId: "ns-1",
        issueKey: "PLAT-1",
      },
      (issue) => applyIssuePatchFields(issue, { status: "done" })
    );

    expect(queryClient.getQueryData<Issue>(byIdOptions.queryKey)?.status).toBe(
      "done"
    );
    expect(queryClient.getQueryData<Issue>(byKeyOptions.queryKey)?.status).toBe(
      "done"
    );
    expect(
      queryClient.getQueryData<PartialIssuePage>(listOptions.queryKey)?.items[0]
        ?.status
    ).toBe("done");

    rollbackIssueCaches(queryClient, snapshot);

    expect(queryClient.getQueryData(byIdOptions.queryKey)).toEqual(original);
    expect(queryClient.getQueryData(byKeyOptions.queryKey)).toEqual(original);
    expect(queryClient.getQueryData(listOptions.queryKey)).toEqual(page);
  });

  it("commits the server issue into detail and list caches", () => {
    const queryClient = new QueryClient();
    const listOptions = v1ProjectsIssuesGetOptions({
      path: { id: "project-1" },
      query: { page_size: 100 },
    });
    const byIdOptions = v1IssueGetOptions({ path: { id: "issue-1" } });
    queryClient.setQueryData(
      listOptions.queryKey,
      issuePage([partialIssue({ status: "open", priority: "normal" })])
    );
    queryClient.setQueryData(
      byIdOptions.queryKey,
      detailIssue({ status: "open" })
    );

    const snapshot = snapshotAndPatchIssueCaches(
      queryClient,
      { issueId: "issue-1", projectId: "project-1" },
      (issue) => applyIssuePatchFields(issue, { status: "done" })
    );
    const updated = detailIssue({ status: "done", priority: "high" });

    commitIssueCaches(queryClient, snapshot, updated, {
      status: updated.status,
      priority: updated.priority,
    });

    expect(queryClient.getQueryData(byIdOptions.queryKey)).toEqual(updated);
    expect(
      queryClient.getQueryData<PartialIssuePage>(listOptions.queryKey)?.items[0]
    ).toMatchObject({ status: "done", priority: "high", title: "Move me" });
  });
});
