import { describe, expect, it } from "vitest";

import { mapIssue } from "./api";

describe("mapIssue", () => {
  it("copies the namespace slug for work-item links", () => {
    expect(
      mapIssue({
        id: "issue-1",
        key: "PLAT-12",
        title: "Charge hours",
        project: { id: "project-1" },
        namespace: { slug: "platform" },
      })
    ).toEqual({
      id: "issue-1",
      key: "PLAT-12",
      title: "Charge hours",
      projectId: "project-1",
      namespaceSlug: "platform",
    });
  });

  it("omits namespaceSlug when the issue has no namespace", () => {
    expect(
      mapIssue({
        id: "issue-1",
        key: "PLAT-12",
        title: "Charge hours",
      }).namespaceSlug
    ).toBeUndefined();
  });
});
