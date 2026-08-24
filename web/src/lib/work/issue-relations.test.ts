import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it } from "vitest";

import {
  ISSUE_RELATION_KIND_SUBTASK_OF,
  editableIssueRelationKinds,
  filterAvailableRelatedIssues,
  issueRelationDisplayKind,
  issueRelationInvalidationKeys,
  issueRelationKindLabel,
  issueRelationKindSelectValues,
  relatedIssueCatalogQueryOptions,
  relatedIssueIds,
  relatedIssueWorkPath,
  relationKindPatch,
  visibleIssueRelations,
} from "./issue-relations";

import { v1IssueRelationsGetOptions } from "@/lib/api/query-options";
import type { IssueRelation, PartialIssue } from "@/lib/api/types";

function partialIssue(
  overrides: Partial<PartialIssue> & Pick<PartialIssue, "id" | "key">
): PartialIssue {
  return {
    numeric_id: 1,
    kind: "task",
    title: "Related work",
    status: "open",
    priority: "normal",
    assignees: [],
    reviewers: [],
    labels: [],
    created_at: "2026-08-01T00:00:00Z",
    updated_at: "2026-08-01T00:00:00Z",
    ...overrides,
  };
}

function relation(
  overrides: Partial<IssueRelation> &
    Pick<IssueRelation, "id" | "kind" | "direction">
): IssueRelation {
  return {
    related: partialIssue({ id: "issue-2", key: "PLAT-2" }),
    created_at: "2026-08-01T00:00:00Z",
    ...overrides,
  };
}

describe("issue relation kind helpers", () => {
  it("excludes subtask of and depends on from editable kinds", () => {
    expect(editableIssueRelationKinds).not.toContain(
      ISSUE_RELATION_KIND_SUBTASK_OF
    );
    expect(editableIssueRelationKinds).not.toContain("depends on");
    expect(editableIssueRelationKinds).toEqual([
      "blocked by",
      "blocks",
      "duplicated by",
      "duplicates",
      "related to",
    ]);
  });

  it("keeps outgoing kinds as-is and inverts incoming kinds", () => {
    expect(issueRelationDisplayKind("blocks", "outgoing")).toBe("blocks");
    expect(issueRelationDisplayKind("blocks", "incoming")).toBe("blocked by");
    expect(issueRelationDisplayKind("blocked by", "incoming")).toBe("blocks");
    expect(issueRelationDisplayKind("duplicates", "incoming")).toBe(
      "duplicated by"
    );
    expect(issueRelationDisplayKind("duplicated by", "incoming")).toBe(
      "duplicates"
    );
    expect(issueRelationDisplayKind("related to", "incoming")).toBe(
      "related to"
    );
  });

  it("treats depends on as blocked by for display", () => {
    expect(issueRelationDisplayKind("depends on", "outgoing")).toBe(
      "blocked by"
    );
    expect(issueRelationDisplayKind("depends on", "incoming")).toBe("blocks");
  });

  it("capitalizes kind labels for display", () => {
    expect(issueRelationKindLabel("blocked by")).toBe("Blocked by");
    expect(issueRelationKindLabel("blocks")).toBe("Blocks");
  });

  it("offers only editable kinds in the kind select", () => {
    expect(issueRelationKindSelectValues()).toEqual([
      ...editableIssueRelationKinds,
    ]);
  });

  it("patches only when the selected kind is a new editable kind", () => {
    expect(relationKindPatch("blocks", "blocks")).toBeNull();
    expect(relationKindPatch("blocked by", "blocks")).toBe("blocks");
    expect(relationKindPatch("blocks", "depends on")).toBeNull();
    expect(
      relationKindPatch("blocks", ISSUE_RELATION_KIND_SUBTASK_OF)
    ).toBeNull();
  });
});

describe("issue relation list helpers", () => {
  it("hides subtask of relations from the editor and inspector lists", () => {
    const items = [
      relation({ id: "rel-1", kind: "blocks", direction: "outgoing" }),
      relation({
        id: "rel-2",
        kind: ISSUE_RELATION_KIND_SUBTASK_OF,
        direction: "outgoing",
      }),
    ];

    expect(visibleIssueRelations(items).map((item) => item.id)).toEqual([
      "rel-1",
    ]);
  });

  it("collects already-related issue ids including hidden subtask of", () => {
    const items = [
      relation({
        id: "rel-1",
        kind: "related to",
        direction: "outgoing",
        related: partialIssue({ id: "issue-2", key: "PLAT-2" }),
      }),
      relation({
        id: "rel-2",
        kind: ISSUE_RELATION_KIND_SUBTASK_OF,
        direction: "outgoing",
        related: partialIssue({ id: "issue-3", key: "PLAT-3" }),
      }),
    ];

    expect([...relatedIssueIds(items)].sort()).toEqual(["issue-2", "issue-3"]);
  });

  it("filters the current issue and already-related ids from the picker", () => {
    const issues = [{ id: "issue-1" }, { id: "issue-2" }, { id: "issue-3" }];

    expect(
      filterAvailableRelatedIssues(issues, "issue-1", new Set(["issue-2"])).map(
        (issue) => issue.id
      )
    ).toEqual(["issue-3"]);
  });

  it("loads the picker catalog from namespace issues, not a single project", () => {
    const options = relatedIssueCatalogQueryOptions("ns-1");
    const queryKey = JSON.stringify(options.queryKey);

    expect(queryKey).toContain("v1NamespacesIssuesGet");
    expect(queryKey).toContain("ns-1");
    expect(queryKey).not.toContain("v1ProjectsIssuesGet");
  });

  it("builds work item paths from the related namespace with a fallback", () => {
    expect(
      relatedIssueWorkPath(
        partialIssue({
          id: "issue-2",
          key: "PLAT-2",
          namespace: { id: "ns-other", name: "Other" },
        }),
        "ns-1"
      )
    ).toBe("/work/ns-other/PLAT-2");
    expect(
      relatedIssueWorkPath(
        partialIssue({ id: "issue-2", key: "PLAT-2" }),
        "ns-1"
      )
    ).toBe("/work/ns-1/PLAT-2");
  });
});

describe("issue relation cache invalidation", () => {
  it("matches relation pages for an issue regardless of page size", () => {
    const queryClient = new QueryClient();
    const editor = v1IssueRelationsGetOptions({
      path: { id: "issue-1" },
      query: { page_size: 100 },
    });
    const inspector = v1IssueRelationsGetOptions({
      path: { id: "issue-1" },
      query: { page_size: 4 },
    });
    const otherIssue = v1IssueRelationsGetOptions({
      path: { id: "issue-2" },
      query: { page_size: 100 },
    });

    queryClient.setQueryData(editor.queryKey, {
      items: [],
      page_info: { has_more: false },
    });
    queryClient.setQueryData(inspector.queryKey, {
      items: [],
      page_info: { has_more: false },
    });
    queryClient.setQueryData(otherIssue.queryKey, {
      items: [],
      page_info: { has_more: false },
    });

    const [relationsKey] = issueRelationInvalidationKeys({
      issueId: "issue-1",
    });

    const matched = queryClient
      .getQueryCache()
      .findAll({ queryKey: relationsKey });

    expect(matched).toHaveLength(2);
    expect(matched.map((query) => query.queryKey)).toEqual(
      expect.arrayContaining([editor.queryKey, inspector.queryKey])
    );
  });

  it("includes issue detail and the related issue keys", () => {
    const keys = issueRelationInvalidationKeys({
      issueId: "issue-1",
      namespaceId: "ns-1",
      issueKey: "PLAT-1",
      related: partialIssue({
        id: "issue-2",
        key: "PLAT-2",
        namespace: { id: "ns-1", name: "Engineering" },
      }),
    });

    expect(keys).toHaveLength(6);
  });
});
