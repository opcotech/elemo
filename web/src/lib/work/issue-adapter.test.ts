import { describe, expect, it } from "vitest";

import type { Issue } from "@/lib/api/types";
import {
  issueStatusLabels,
  issueToWorkItem,
  issuesToWorkItemsWithNamespaces,
  mapIssueStatus,
  mapWorkStatusToIssueStatus,
} from "@/lib/work/issue-adapter";

const baseIssue: Issue = {
  id: "issue-1",
  key: "PLAT-7",
  numeric_id: 7,
  kind: "story",
  title: "Ship work surface",
  description: "Wire real issues into project work.",
  status: "in progress",
  priority: "highest",
  resolution: "none",
  reported_by: {
    id: "user-1",
    first_name: "Test",
    last_name: "User",
    picture: null,
  },
  assignees: [
    { id: "user-2", first_name: "Ada", last_name: "Lovelace", picture: null },
    { id: "user-3", first_name: "Grace", last_name: "Hopper", picture: null },
  ],
  reviewers: [
    {
      id: "user-4",
      first_name: "Katherine",
      last_name: "Johnson",
      picture: null,
    },
  ],
  labels: [{ id: "label-1", name: "frontend" }],
  parent: {
    id: "issue-0",
    key: "PLAT-1",
    numeric_id: 1,
    kind: "epic",
    title: "Parent epic",
    status: "open",
    priority: "high",
    assignees: [],
    reviewers: [],
    labels: [],
    created_at: "2026-08-01T00:00:00Z",
    updated_at: "2026-08-10T00:00:00Z",
    namespace: {
      id: "ns-1",
      slug: "engineering",
      name: "Engineering",
    },
  },
  links: [{ url: "https://example.com/spec", label: "Design spec" }],
  project: {
    id: "proj-1",
    key: "PLAT",
    name: "Platform",
    status: "active",
  },
  namespace: {
    id: "ns-1",
    slug: "engineering",
    name: "Engineering",
  },
  comment_count: 0,
  attachment_count: 0,
  watcher_count: 0,
  relation_count: 0,
  due_date: "2026-08-20T00:00:00Z",
  start_date: null,
  created_at: "2026-08-01T00:00:00Z",
  updated_at: "2026-08-10T00:00:00Z",
};

describe("issue-adapter", () => {
  it("maps API status into work board values", () => {
    expect(mapIssueStatus("open")).toBe("backlog");
    expect(mapIssueStatus("in progress")).toBe("in progress");
    expect(mapIssueStatus("review")).toBe("in review");
    expect(mapIssueStatus("closed")).toBe("closed");
    expect(issueStatusLabels.open).toBe("Backlog");
    expect(issueStatusLabels.review).toBe("In review");
  });

  it("maps work board status back to API values", () => {
    expect(mapWorkStatusToIssueStatus("backlog")).toBe("open");
    expect(mapWorkStatusToIssueStatus("in progress")).toBe("in progress");
    expect(mapWorkStatusToIssueStatus("in review")).toBe("review");
    expect(mapWorkStatusToIssueStatus("closed")).toBe("closed");
  });

  it("adapts issues into work items using the API composite key", () => {
    const item = issueToWorkItem(baseIssue, {
      namespaceId: "ns-1",
      projectId: "proj-1",
    });

    expect(item).toMatchObject({
      dataSource: "api",
      id: "issue-1",
      key: "PLAT-7",
      title: "Ship work surface",
      summary: "Wire real issues into project work.",
      namespaceId: "ns-1",
      projectId: "proj-1",
      namespace: { id: "ns-1", name: "Engineering" },
      project: { id: "proj-1", key: "PLAT", name: "Platform" },
      status: "in progress",
      priority: "highest",
      assigneeIds: ["user-2", "user-3"],
      reviewerIds: ["user-4"],
      assignees: [
        { id: "user-2", name: "Ada Lovelace", picture: null },
        { id: "user-3", name: "Grace Hopper", picture: null },
      ],
      reviewers: [{ id: "user-4", name: "Katherine Johnson", picture: null }],
      assigneeId: "user-2",
      creatorId: "user-1",
      labelIds: ["label-1"],
      labels: [{ id: "label-1", name: "frontend" }],
      rank: 7,
      startDate: null,
      dueDate: "2026-08-20T00:00:00Z",
      kind: "story",
      resolution: "none",
      parent: {
        id: "issue-0",
        key: "PLAT-1",
        title: "Parent epic",
        namespaceId: "ns-1",
        namespaceSlug: "engineering",
      },
      links: [{ url: "https://example.com/spec", label: "Design spec" }],
    });
  });

  it("defaults missing assignment and label arrays", () => {
    const item = issueToWorkItem(
      {
        id: baseIssue.id,
        key: baseIssue.key,
        numeric_id: baseIssue.numeric_id,
        kind: baseIssue.kind,
        title: baseIssue.title,
        status: baseIssue.status,
        priority: baseIssue.priority,
        created_at: baseIssue.created_at,
        updated_at: baseIssue.updated_at,
      },
      { namespaceId: "ns-1" }
    );

    expect(item.assigneeIds).toEqual([]);
    expect(item.reviewerIds).toEqual([]);
    expect(item.assignees).toEqual([]);
    expect(item.reviewers).toEqual([]);
    expect(item.labelIds).toEqual([]);
    expect(item.labels).toEqual([]);
    expect(item.summary).toBe("");
    expect(item.assigneeId).toBeNull();
    expect(item.namespace).toBeUndefined();
    expect(item.project).toBeUndefined();
    expect(item.parent).toBeUndefined();
    expect(item.links).toEqual([]);
    expect(item.kind).toBe("story");
    expect(item.resolution).toBeUndefined();
    expect(item.creatorId).toBe("");
  });

  it("maps list issue reporters onto work items", () => {
    const item = issueToWorkItem(
      {
        id: baseIssue.id,
        key: baseIssue.key,
        numeric_id: baseIssue.numeric_id,
        kind: baseIssue.kind,
        title: baseIssue.title,
        status: baseIssue.status,
        priority: baseIssue.priority,
        created_at: baseIssue.created_at,
        updated_at: baseIssue.updated_at,
        reported_by: {
          id: "user-reporter",
          first_name: "Reporter",
          last_name: "User",
          picture: null,
        },
      },
      { namespaceId: "ns-1" }
    );

    expect(item.creatorId).toBe("user-reporter");
  });

  it("maps a cleared parent to null", () => {
    const item = issueToWorkItem(
      {
        ...baseIssue,
        parent: null,
        resolution: "fixed",
        links: [],
      },
      { namespaceId: "ns-1" }
    );

    expect(item.parent).toBeNull();
    expect(item.resolution).toBe("fixed");
    expect(item.links).toEqual([]);
  });

  it("fills hierarchical slugs from reachable namespaces for cross-namespace lists", () => {
    const items = issuesToWorkItemsWithNamespaces(
      [
        baseIssue,
        {
          ...baseIssue,
          id: "issue-2",
          key: "OPS-3",
          namespace: {
            id: "ns-2",
            slug: "ops",
            name: "Operations",
          },
        },
      ],
      [
        {
          id: "ns-1",
          slug: "engineering",
          organizationId: "org-1",
          organizationSlug: "acme",
        },
        {
          id: "ns-2",
          slug: "ops",
          organizationId: "org-2",
          organizationSlug: "globex",
        },
      ]
    );

    expect(items[0]).toMatchObject({
      organizationSlug: "acme",
      namespaceSlug: "engineering",
      organizationId: "org-1",
    });
    expect(items[1]).toMatchObject({
      organizationSlug: "globex",
      namespaceSlug: "ops",
      organizationId: "org-2",
    });
  });
});
