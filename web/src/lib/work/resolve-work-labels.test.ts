import { describe, expect, it } from "vitest";

import {
  labelsFromIds,
  mergeWorkLabels,
  resolveWorkLabels,
} from "@/lib/work/resolve-work-labels";

describe("resolveWorkLabels", () => {
  it("maps label ids to catalog names", () => {
    expect(
      resolveWorkLabels(
        ["label-1", "label-2"],
        [
          { id: "label-1", name: "frontend" },
          { id: "label-2", name: "api" },
        ]
      )
    ).toEqual([
      { id: "label-1", name: "frontend" },
      { id: "label-2", name: "api" },
    ]);
  });

  it("falls back to the id when the catalog has no match", () => {
    expect(resolveWorkLabels(["frontend"], [])).toEqual([
      { id: "frontend", name: "frontend" },
    ]);
  });

  it("matches mock label ids that are already names", () => {
    expect(
      resolveWorkLabels(["frontend"], [{ id: "label-1", name: "frontend" }])
    ).toEqual([{ id: "frontend", name: "frontend" }]);
  });
});

describe("mergeWorkLabels", () => {
  it("keeps selected names without waiting for a catalog", () => {
    expect(mergeWorkLabels([{ id: "label-1", name: "frontend" }])).toEqual([
      { id: "label-1", name: "frontend" },
    ]);
  });

  it("appends catalog labels that are not already selected", () => {
    expect(
      mergeWorkLabels(
        [{ id: "label-1", name: "frontend" }],
        [
          { id: "label-1", name: "ignored" },
          { id: "label-2", name: "api" },
        ]
      )
    ).toEqual([
      { id: "label-1", name: "frontend" },
      { id: "label-2", name: "api" },
    ]);
  });

  it("replaces an id placeholder with the catalog name", () => {
    expect(
      mergeWorkLabels(
        [{ id: "label-2", name: "label-2" }],
        [{ id: "label-2", name: "api" }]
      )
    ).toEqual([{ id: "label-2", name: "api" }]);
  });
});

describe("labelsFromIds", () => {
  it("prefers catalog names over id placeholders from the issue", () => {
    expect(
      labelsFromIds(
        ["label-1", "label-2"],
        [{ id: "label-1", name: "frontend" }],
        [
          { id: "label-1", name: "frontend" },
          { id: "label-2", name: "api" },
        ]
      )
    ).toEqual([
      { id: "label-1", name: "frontend" },
      { id: "label-2", name: "api" },
    ]);
  });
});
