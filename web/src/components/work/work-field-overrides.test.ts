import { describe, expect, it } from "vitest";

import type { WorkFieldOverride } from "./work-field-overrides";
import { rollbackWorkFieldOverride } from "./work-field-overrides";

describe("rollbackWorkFieldOverride", () => {
  it("removes only the failed fields and keeps the rest", () => {
    const overrides = new Map<string, WorkFieldOverride>([
      [
        "issue-1",
        {
          status: "done",
          startDate: "2026-08-01T12:00:00.000Z",
          dueDate: "2026-08-08T12:00:00.000Z",
        },
      ],
    ]);

    const next = rollbackWorkFieldOverride(overrides, "issue-1", ["status"]);

    expect(next.get("issue-1")).toEqual({
      startDate: "2026-08-01T12:00:00.000Z",
      dueDate: "2026-08-08T12:00:00.000Z",
    });
    expect(overrides.get("issue-1")?.status).toBe("done");
  });

  it("deletes the issue entry when no overrides remain", () => {
    const overrides = new Map<string, WorkFieldOverride>([
      ["issue-1", { priority: "high" }],
    ]);

    const next = rollbackWorkFieldOverride(overrides, "issue-1", ["priority"]);

    expect(next.has("issue-1")).toBe(false);
    expect(next.size).toBe(0);
  });

  it("returns the same map when the issue has no matching fields", () => {
    const overrides = new Map<string, WorkFieldOverride>([
      ["issue-1", { status: "done" }],
    ]);

    expect(
      rollbackWorkFieldOverride(overrides, "issue-1", ["startDate", "dueDate"])
    ).toBe(overrides);
    expect(rollbackWorkFieldOverride(overrides, "missing", ["status"])).toBe(
      overrides
    );
  });
});
