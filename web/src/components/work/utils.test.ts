import { describe, expect, it } from "vitest";

import {
  createTimelineScale,
  paginate,
  parseSort,
  selectedWorkId,
  timelinePosition,
} from "./utils";

import type { TimelineEntry } from "@/lib/mock-data";

const entries = [
  {
    dataSource: "mock",
    id: "entry-one",
    workItemId: "lmo-101",
    namespaceId: "namespace-one",
    projectId: null,
    title: "First",
    startAt: "2026-08-05T09:00:00.000Z",
    endAt: "2026-08-12T17:00:00.000Z",
    kind: "work",
  },
  {
    dataSource: "mock",
    id: "entry-two",
    workItemId: "ops-301",
    namespaceId: "namespace-one",
    projectId: null,
    title: "Milestone",
    startAt: "2026-08-17T09:00:00.000Z",
    endAt: "2026-08-17T09:00:00.000Z",
    kind: "milestone",
  },
] as const satisfies readonly TimelineEntry[];

describe("work projection utilities", () => {
  it("derives timeline headers and positions from entry dates", () => {
    const scale = createTimelineScale(entries);

    expect(scale?.ticks.map((tick) => tick.toISOString())).toEqual([
      "2026-08-03T00:00:00.000Z",
      "2026-08-10T00:00:00.000Z",
      "2026-08-17T00:00:00.000Z",
      "2026-08-24T00:00:00.000Z",
    ]);
    expect(scale && timelinePosition(entries[0], scale)).toMatchObject({
      left: expect.closeTo(8.48, 1),
      width: expect.closeTo(26.19, 1),
    });
    expect(scale && timelinePosition(entries[1], scale).width).toBe(0);
  });

  it("clamps pagination to a valid page", () => {
    expect(paginate([1, 2, 3, 4, 5], 4, 2)).toEqual({
      items: [5],
      page: 3,
      pageCount: 3,
    });
  });

  it("extracts selected work ids from URL-owned selection tokens", () => {
    expect(selectedWorkId("work:lmo-101")).toBe("lmo-101");
    expect(selectedWorkId("work:ops-301")).toBe("ops-301");
    expect(selectedWorkId("document:document-1")).toBeUndefined();
    expect(selectedWorkId(undefined)).toBeUndefined();
  });

  it("normalizes projection sort tokens", () => {
    expect(parseSort("priority:desc")).toBe("priority:desc");
    expect(parseSort("rank")).toBe("rank:asc");
    expect(parseSort("")).toBe("rank:asc");
  });
});
