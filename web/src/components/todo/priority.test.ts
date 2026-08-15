import { describe, expect, it } from "vitest";

import {
  todoPriorities,
  todoPriorityLabels,
  todoPriorityToneClassName,
} from "./priority";

describe("todo priority", () => {
  it("exposes human labels for every todo priority", () => {
    expect(todoPriorities).toEqual([
      "normal",
      "important",
      "urgent",
      "critical",
    ]);
    expect(todoPriorityLabels).toEqual({
      normal: "Normal",
      important: "Important",
      urgent: "Urgent",
      critical: "Critical",
    });
  });

  it("uses the planned tone classes", () => {
    expect(todoPriorityToneClassName).toEqual({
      normal: "text-muted-foreground",
      important: "text-primary",
      urgent: "text-warning-on-subtle",
      critical: "text-destructive",
    });
  });
});
