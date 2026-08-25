import type { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/errors";
import type { AccessibleNamespace, Issue } from "@/lib/api/types";
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
    slug: "product",
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

const accessibleNamespace: AccessibleNamespace = {
  id: "namespace-product",
  slug: "product",
  name: "Product",
  created_at: "2026-08-01T00:00:00Z",
  updated_at: "2026-08-10T00:00:00Z",
  organization: {
    id: "organization-acme",
    slug: "acme",
    name: "Acme",
  },
};

function createQueryClient(result: unknown) {
  return {
    fetchQuery: vi.fn((options: { queryKey: readonly unknown[] }) => {
      if (result instanceof Error) {
        return Promise.reject(result);
      }
      const key = JSON.stringify(options.queryKey);
      if (key.includes("v1NamespaceGet")) {
        return Promise.resolve(accessibleNamespace);
      }
      return Promise.resolve(result);
    }),
    setQueryData: vi.fn(),
  } as unknown as QueryClient;
}

describe("loadWorkItemPage", () => {
  it("loads an issue by organization slug, namespace slug, and issue key", async () => {
    const queryClient = createQueryClient(liveIssue);

    await expect(
      loadWorkItemPage(queryClient, "acme", "product", "LMO-101")
    ).resolves.toMatchObject({
      source: "api",
      issue: liveIssue,
      organizationSlug: "acme",
      namespaceSlug: "product",
      organizationId: "organization-acme",
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

  it("rejects xid route segments without calling the API", async () => {
    const queryClient = createQueryClient(liveIssue);

    await expect(
      loadWorkItemPage(
        queryClient,
        "9bsv0s46s6s002p9ltq0",
        "product",
        "LMO-101"
      )
    ).rejects.toMatchObject({ isNotFound: true });
    expect(queryClient.fetchQuery).not.toHaveBeenCalled();
  });

  it("throws not-found when the API 404s", async () => {
    const queryClient = createQueryClient(new ApiError(404, "missing"));

    await expect(
      loadWorkItemPage(queryClient, "acme", "product", "LMO-101")
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("does not substitute fixtures for permission errors", async () => {
    const queryClient = createQueryClient(new ApiError(403, "denied"));

    await expect(
      loadWorkItemPage(queryClient, "acme", "product", "LMO-101")
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("propagates unexpected API failures", async () => {
    const error = new ApiError(500, "boom");
    const queryClient = createQueryClient(error);

    await expect(
      loadWorkItemPage(queryClient, "acme", "product", "LMO-101")
    ).rejects.toBe(error);
  });

  it("maps validation errors to not-found for noncanonical keys", async () => {
    const queryClient = createQueryClient(new ApiError(400, "invalid"));

    await expect(
      loadWorkItemPage(queryClient, "acme", "product", "lmo-101")
    ).rejects.toMatchObject({ isNotFound: true });
    expect(queryClient.fetchQuery).not.toHaveBeenCalled();
  });
});
