import type { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/errors";
import type { Issue } from "@/lib/api/types";
import { loadWorkItemPage } from "@/lib/work/load-work-item";

const liveIssue: Issue = {
  id: "issue-live",
  key: "LMO-101",
  numeric_id: 101,
  kind: "task",
  title: "Live issue that shares a fixture key",
  description: "API should win over fixtures.",
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
  project: {
    id: "project-web",
    key: "LMO",
    name: "Web",
    status: "active",
  },
  namespace: {
    id: "namespace-product",
    name: "Product",
  },
  comment_count: 0,
  attachment_count: 0,
  watcher_count: 0,
  relation_count: 0,
  due_date: null,
  start_date: null,
  created_at: "2026-08-01T00:00:00Z",
  updated_at: "2026-08-10T00:00:00Z",
};

function createQueryClient(result: unknown) {
  return {
    fetchQuery: vi.fn(() =>
      result instanceof Error ? Promise.reject(result) : Promise.resolve(result)
    ),
  } as unknown as QueryClient;
}

describe("loadWorkItemPage", () => {
  it("prefers a live API issue over a fixture with the same key", async () => {
    const queryClient = createQueryClient(liveIssue);

    await expect(
      loadWorkItemPage(queryClient, "namespace-product", "LMO-101")
    ).resolves.toMatchObject({
      source: "api",
      issue: liveIssue,
      namespaceId: "namespace-product",
      issueKey: "LMO-101",
      item: {
        dataSource: "api",
        id: "issue-live",
        key: "LMO-101",
        title: "Live issue that shares a fixture key",
      },
    });
  });

  it("falls back to a fixture only when the API returns 404", async () => {
    const queryClient = createQueryClient(new ApiError(404, "missing"));

    const data = await loadWorkItemPage(
      queryClient,
      "namespace-product",
      "LMO-101"
    );

    expect(data).toMatchObject({
      source: "mock",
      item: {
        dataSource: "mock",
        key: "LMO-101",
        namespaceId: "namespace-product",
      },
    });
  });

  it("throws not-found when the API 404s and no fixture matches", async () => {
    const queryClient = createQueryClient(new ApiError(404, "missing"));

    await expect(
      loadWorkItemPage(queryClient, "namespace-product", "MISSING-1")
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("does not substitute fixtures for permission errors", async () => {
    const queryClient = createQueryClient(new ApiError(403, "denied"));

    await expect(
      loadWorkItemPage(queryClient, "namespace-product", "LMO-101")
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("propagates unexpected API failures", async () => {
    const error = new ApiError(500, "boom");
    const queryClient = createQueryClient(error);

    await expect(
      loadWorkItemPage(queryClient, "namespace-product", "LMO-101")
    ).rejects.toBe(error);
  });
});
