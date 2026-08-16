import { describe, expect, it } from "vitest";

import {
  issueSelectSearchText,
  issueToSelectOption,
} from "./issue-select-option";

const issue = {
  id: "issue-1",
  key: "PLAT-7",
  title: "Ship work surface",
  kind: "bug" as const,
  status: "in progress" as const,
  priority: "highest" as const,
  project: {
    id: "project-1",
    key: "PLAT",
    name: "Platform",
    status: "active" as const,
  },
};

describe("issue select options", () => {
  it("puts kind, project, status, and priority into search text", () => {
    expect(issueSelectSearchText(issue)).toBe(
      "bug Bug in progress In progress highest Highest"
    );
  });

  it("keeps kind-only search text when the issue has no project", () => {
    expect(
      issueSelectSearchText({
        kind: "bug",
        status: "in progress",
        priority: "highest",
      })
    ).toBe("bug Bug in progress In progress highest Highest");
  });

  it("builds a selectable option with kind first in details", () => {
    const option = issueToSelectOption(issue);

    expect(option.value).toBe("issue-1");
    expect(option.title).toBe("PLAT-7 Ship work surface");
    expect(option.searchText).toContain("bug");
    expect(option.searchText).toContain("In progress");
    expect(option.title).toContain("PLAT");
    expect(option.details).toBeDefined();
  });
});
