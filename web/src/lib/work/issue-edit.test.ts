import { describe, expect, it } from "vitest";

import type { PartialProject } from "@/lib/client";
import {
  assignmentIdsEqual,
  matchProjectByIssueKey,
  normalizeAssignmentIds,
  parseIssueDescription,
  parseIssueTitle,
} from "@/lib/work/issue-edit";

describe("issue-edit helpers", () => {
  describe("parseIssueTitle", () => {
    it("accepts a valid title", () => {
      expect(parseIssueTitle("  Ship the feature  ")).toEqual({
        ok: true,
        title: "Ship the feature",
      });
    });

    it("rejects titles that are too short", () => {
      const result = parseIssueTitle("ab");
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.error.length).toBeGreaterThan(0);
      }
    });
  });

  describe("parseIssueDescription", () => {
    it("clears empty descriptions to null", () => {
      expect(parseIssueDescription("   ")).toEqual({
        ok: true,
        description: null,
      });
      expect(parseIssueDescription("", "**")).toEqual({
        ok: true,
        description: null,
      });
    });

    it("accepts plain text that meets the minimum length", () => {
      expect(parseIssueDescription("Enough detail here")).toEqual({
        ok: true,
        description: "Enough detail here",
      });
    });

    it("persists markdown when plain text is long enough", () => {
      expect(
        parseIssueDescription("Enough detail here", "**Enough detail here**")
      ).toEqual({
        ok: true,
        description: "**Enough detail here**",
      });
    });

    it("rejects short plain text even when markdown markers are longer", () => {
      const result = parseIssueDescription("ab", "**ab**");
      expect(result.ok).toBe(false);
    });

    it("accepts the new minimum plain-text length", () => {
      expect(parseIssueDescription("abc")).toEqual({
        ok: true,
        description: "abc",
      });
    });

    it("accepts legacy plain-text strings as the markdown payload", () => {
      expect(
        parseIssueDescription("Wire real issues into project work.")
      ).toEqual({
        ok: true,
        description: "Wire real issues into project work.",
      });
    });
  });

  describe("matchProjectByIssueKey", () => {
    const projects: PartialProject[] = [
      {
        id: "p1",
        key: "LMO",
        name: "Elemo",
        status: "active",
      },
      {
        id: "p2",
        key: "LMOX",
        name: "Elemo Extended",
        status: "active",
      },
    ];

    it("matches the longest project key prefix", () => {
      expect(matchProjectByIssueKey("LMOX-12", projects)?.id).toBe("p2");
      expect(matchProjectByIssueKey("LMO-7", projects)?.id).toBe("p1");
    });

    it("returns undefined when no project matches", () => {
      expect(matchProjectByIssueKey("ZZZ-1", projects)).toBeUndefined();
    });
  });

  describe("normalizeAssignmentIds", () => {
    it("returns an empty array for empty selection", () => {
      expect(normalizeAssignmentIds([])).toEqual([]);
    });

    it("returns an empty array when ids are missing", () => {
      expect(normalizeAssignmentIds(undefined)).toEqual([]);
      expect(normalizeAssignmentIds(null)).toEqual([]);
    });

    it("keeps multiple ids in first-seen order and drops blanks/dupes", () => {
      expect(
        normalizeAssignmentIds(["user-2", "  ", "user-1", "user-2", "user-3"])
      ).toEqual(["user-2", "user-1", "user-3"]);
    });
  });

  describe("assignmentIdsEqual", () => {
    it("compares assignment lists without regard to order", () => {
      expect(assignmentIdsEqual(["a", "b"], ["b", "a"])).toBe(true);
      expect(assignmentIdsEqual(["a"], ["a", "b"])).toBe(false);
      expect(assignmentIdsEqual([], [])).toBe(true);
    });
  });
});
