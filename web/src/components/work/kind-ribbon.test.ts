import { describe, expect, it } from "vitest";

import {
  issueKindLabels,
  issueKindToneClassName,
  issueKinds,
} from "./kind-ribbon";

describe("KindRibbon", () => {
  it("exposes human labels for every issue kind", () => {
    expect(issueKinds).toEqual(["epic", "story", "task", "bug"]);
    expect(issueKindLabels).toEqual({
      epic: "Epic",
      story: "Story",
      task: "Task",
      bug: "Bug",
    });
  });

  it("uses the planned tone classes", () => {
    expect(issueKindToneClassName).toEqual({
      epic: "text-purple-500",
      story: "text-success",
      task: "text-primary",
      bug: "text-destructive",
    });
  });
});
