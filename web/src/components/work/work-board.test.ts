import { describe, expect, it } from "vitest";

import { resolveBoardDropColumn } from "./work-board";

import type { WorkItem } from "@/lib/work/model";

function workItem(overrides: Partial<WorkItem> = {}): WorkItem {
  return {
    dataSource: "api",
    id: "issue-1",
    key: "PLAT-1",
    title: "Move me",
    summary: "Move me",
    namespaceId: "ns-1",
    projectId: "project-1",
    status: "backlog",
    priority: "normal",
    assigneeIds: [],
    reviewerIds: [],
    assigneeId: null,
    creatorId: "user-1",
    labelIds: [],
    rank: 1,
    dueDate: null,
    startDate: null,
    createdAt: "2026-08-01T00:00:00Z",
    updatedAt: "2026-08-01T00:00:00Z",
    ...overrides,
  };
}

describe("resolveBoardDropColumn", () => {
  const item = workItem();
  const columns = new Map<string, readonly WorkItem[]>([
    ["backlog", [item]],
    ["done", []],
  ]);

  it("uses the column droppable when the card is still in its source column", () => {
    expect(resolveBoardDropColumn(columns, item.id, "column:done")).toBe(
      "done"
    );
  });

  it("uses the hovered card's column when dropping onto another item", () => {
    const doneItem = workItem({ id: "issue-2", key: "PLAT-2", status: "done" });
    const populated = new Map<string, readonly WorkItem[]>([
      ["backlog", [item]],
      ["done", [doneItem]],
    ]);

    expect(resolveBoardDropColumn(populated, item.id, doneItem.id)).toBe(
      "done"
    );
  });

  it("falls back to the active card's column when the drop target is unknown", () => {
    expect(resolveBoardDropColumn(columns, item.id, "unknown")).toBe("backlog");
  });
});
