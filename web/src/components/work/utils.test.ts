import { format } from "date-fns";
import { describe, expect, it, vi } from "vitest";

import {
  applyTimelineDrag,
  calendarDateToUtcNoonIso,
  createTimelineScale,
  dateLabel,
  formatDateTime,
  formatTargetDate,
  paginate,
  parseSort,
  selectedWorkId,
  shiftIsoByUtcDays,
  timelinePosition,
  timelineRangeFromSpan,
  timelineRangeLabel,
  utcDayDelta,
  utcIsoToCalendarDate,
  workItemDatesFromTimelineRange,
  workItemPath,
  workItemTimelineRange,
  workItemUrl,
  workItemsToTimelineEntries,
} from "./utils";

import type { TimelineEntry } from "@/lib/mock-data";
import type { WorkItem } from "@/lib/work/model";

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
  it("stores a picked calendar day at UTC noon instead of local midnight", () => {
    expect(calendarDateToUtcNoonIso(new Date(2026, 7, 20))).toBe(
      "2026-08-20T12:00:00.000Z"
    );
    expect(calendarDateToUtcNoonIso(new Date(2026, 7, 20, 23, 45, 12))).toBe(
      "2026-08-20T12:00:00.000Z"
    );
    expect(calendarDateToUtcNoonIso(new Date(2026, 7, 20))).not.toBe(
      new Date(2026, 7, 20).toISOString()
    );
  });

  it("reads stored UTC timestamps back as the UTC calendar day", () => {
    expect(utcIsoToCalendarDate("2026-08-20T00:00:00.000Z")).toEqual(
      new Date(2026, 7, 20)
    );
    expect(utcIsoToCalendarDate("2026-08-20T12:00:00.000Z")).toEqual(
      new Date(2026, 7, 20)
    );
    expect(utcIsoToCalendarDate("2026-08-19T20:00:00.000Z")).toEqual(
      new Date(2026, 7, 19)
    );
  });

  it("formats target dates like the DatePicker PPP style", () => {
    expect(formatTargetDate(null)).toBe("Unscheduled");
    expect(formatTargetDate("not-a-date")).toBe("Unscheduled");
    expect(formatTargetDate("2026-08-20T00:00:00.000Z")).toBe(
      format(new Date(2026, 7, 20), "PPP")
    );
  });

  it("formats inspector and card dates on the same UTC calendar day", () => {
    const utcMidnight = "2026-08-20T00:00:00.000Z";
    const utcNoon = "2026-08-20T12:00:00.000Z";
    const shiftedLocalMidnight = "2026-08-19T20:00:00.000Z";

    expect(formatTargetDate(utcMidnight)).toBe(
      format(new Date(2026, 7, 20), "PPP")
    );
    expect(formatTargetDate(utcNoon)).toBe(
      format(new Date(2026, 7, 20), "PPP")
    );
    expect(dateLabel(utcMidnight)).toBe(dateLabel(utcNoon));
    expect(dateLabel(shiftedLocalMidnight)).toBe(
      dateLabel("2026-08-19T00:00:00.000Z")
    );
    expect(formatTargetDate(shiftedLocalMidnight)).toBe(
      formatTargetDate("2026-08-19T00:00:00.000Z")
    );
    expect(formatTargetDate(shiftedLocalMidnight)).not.toBe(
      formatTargetDate(utcMidnight)
    );
  });
  it("formats timestamps with date and time", () => {
    expect(formatDateTime(null)).toBe("—");
    expect(formatDateTime("not-a-date")).toBe("—");
    expect(formatDateTime("2026-08-10T15:30:00.000Z")).toMatch(/Aug/);
  });
  it("pads the timeline one month before the earliest start and three months after the latest due", () => {
    expect(createTimelineScale([])).toBeUndefined();

    const scale = createTimelineScale(entries);

    expect(scale?.start).toBe(Date.parse("2026-06-29T00:00:00.000Z"));
    expect(scale?.end).toBe(Date.parse("2026-11-23T00:00:00.000Z"));
    expect(scale?.ticks[0]?.toISOString()).toBe("2026-06-29T00:00:00.000Z");
    expect(scale?.ticks.at(-1)?.toISOString()).toBe("2026-11-16T00:00:00.000Z");
    expect(scale?.ticks).toHaveLength(21);
    expect(scale && timelinePosition(entries[0], scale)).toMatchObject({
      left: expect.closeTo(25.17, 1),
      width: expect.closeTo(4.76, 1),
    });
    expect(scale && timelinePosition(entries[1], scale).width).toBe(0);
  });

  it("aligns bars that share a calendar date regardless of time of day", () => {
    const scale = createTimelineScale(entries);
    expect(scale).toBeDefined();
    if (!scale) {
      return;
    }

    const ending = timelinePosition(
      {
        startAt: "2026-08-05T09:00:00.000Z",
        endAt: "2026-08-12T17:00:00.000Z",
        kind: "work",
      },
      scale
    );
    const starting = timelinePosition(
      {
        startAt: "2026-08-12T09:00:00.000Z",
        endAt: "2026-08-20T17:00:00.000Z",
        kind: "work",
      },
      scale
    );

    expect(ending.right).toBe(starting.left);
    expect(ending.right).toBe(timelinePosition(entries[0], scale).right);
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

  it("builds hierarchical work item paths", () => {
    expect(
      workItemPath({
        organizationSlug: "acme",
        namespaceSlug: "product",
        key: "LMO-101",
      })
    ).toBe("/work/acme/product/LMO-101");
  });

  it("builds an absolute work item URL from the current origin", () => {
    vi.stubGlobal("window", {
      location: { origin: "https://elemo.test" },
    });
    try {
      expect(
        workItemUrl({
          organizationSlug: "acme",
          namespaceSlug: "product",
          key: "LMO-101",
        })
      ).toBe("https://elemo.test/work/acme/product/LMO-101");
    } finally {
      vi.unstubAllGlobals();
    }
  });

  it("normalizes projection sort tokens", () => {
    expect(parseSort("priority:desc")).toBe("priority:desc");
    expect(parseSort("rank")).toBe("rank:asc");
    expect(parseSort("")).toBe("rank:asc");
  });

  it("labels timeline bars with the start-to-due range", () => {
    expect(
      timelineRangeLabel({
        startAt: "2026-08-05T09:00:00.000Z",
        endAt: "2026-08-12T17:00:00.000Z",
        kind: "work",
      })
    ).toBe(
      `${dateLabel("2026-08-05T09:00:00.000Z")} – ${dateLabel("2026-08-12T17:00:00.000Z")}`
    );
    expect(
      timelineRangeLabel({
        startAt: "2026-08-12T17:00:00.000Z",
        endAt: "2026-08-12T17:00:00.000Z",
        kind: "milestone",
      })
    ).toBe(dateLabel("2026-08-12T17:00:00.000Z"));
  });

  it("maps work start and due dates onto the timeline", () => {
    expect(
      workItemTimelineRange({
        startDate: "2026-08-05T09:00:00.000Z",
        dueDate: "2026-08-12T17:00:00.000Z",
      })
    ).toEqual({
      startAt: "2026-08-05T09:00:00.000Z",
      endAt: "2026-08-12T17:00:00.000Z",
      kind: "work",
    });
    expect(
      workItemTimelineRange({
        startDate: null,
        dueDate: "2026-08-12T17:00:00.000Z",
      })
    ).toEqual({
      startAt: "2026-08-12T17:00:00.000Z",
      endAt: "2026-08-12T17:00:00.000Z",
      kind: "milestone",
    });
    expect(
      workItemTimelineRange({
        startDate: "2026-08-05T09:00:00.000Z",
        dueDate: null,
      })
    ).toEqual({
      startAt: "2026-08-05T09:00:00.000Z",
      endAt: "2026-08-05T09:00:00.000Z",
      kind: "milestone",
    });
    expect(
      workItemTimelineRange({ startDate: null, dueDate: null })
    ).toBeUndefined();
    expect(
      workItemTimelineRange({
        startDate: "2026-08-12T17:00:00.000Z",
        dueDate: "2026-08-05T09:00:00.000Z",
      })
    ).toEqual({
      startAt: "2026-08-05T09:00:00.000Z",
      endAt: "2026-08-12T17:00:00.000Z",
      kind: "work",
    });
  });

  it("builds sorted timeline entries from work items", () => {
    const later = workItem({
      id: "later",
      startDate: "2026-08-10T00:00:00.000Z",
      dueDate: "2026-08-12T00:00:00.000Z",
    });
    const earlier = workItem({
      id: "earlier",
      startDate: "2026-08-03T00:00:00.000Z",
      dueDate: "2026-08-08T00:00:00.000Z",
    });
    const unscheduled = workItem({
      id: "unscheduled",
      startDate: null,
      dueDate: null,
    });

    expect(
      workItemsToTimelineEntries([later, unscheduled, earlier]).map(
        (entry) => entry.item.id
      )
    ).toEqual(["earlier", "later"]);
  });

  it("converts pixel movement on the scale into UTC day deltas", () => {
    const scale = {
      start: Date.parse("2026-08-03T00:00:00.000Z"),
      end: Date.parse("2026-08-31T00:00:00.000Z"),
      ticks: [],
    };

    expect(utcDayDelta(10, 280, scale)).toBe(1);
    expect(utcDayDelta(-10, 280, scale)).toBe(-1);
    expect(utcDayDelta(4, 280, scale)).toBe(0);
    expect(utcDayDelta(50, 0, scale)).toBe(0);
  });

  it("shifts ISO timestamps by UTC calendar days and keeps the time of day", () => {
    expect(shiftIsoByUtcDays("2026-08-05T09:00:00.000Z", 2)).toBe(
      "2026-08-07T09:00:00.000Z"
    );
    expect(shiftIsoByUtcDays("2026-08-05T09:00:00.000Z", -3)).toBe(
      "2026-08-02T09:00:00.000Z"
    );
    expect(shiftIsoByUtcDays("2026-08-05T09:00:00.000Z", 0)).toBe(
      "2026-08-05T09:00:00.000Z"
    );
  });

  it("moves or resizes a timeline range without letting start pass due", () => {
    expect(
      applyTimelineDrag({
        startAt: "2026-08-05T09:00:00.000Z",
        endAt: "2026-08-12T17:00:00.000Z",
        mode: "move",
        days: 2,
      })
    ).toEqual({
      startAt: "2026-08-07T09:00:00.000Z",
      endAt: "2026-08-14T17:00:00.000Z",
    });
    expect(
      applyTimelineDrag({
        startAt: "2026-08-05T09:00:00.000Z",
        endAt: "2026-08-12T17:00:00.000Z",
        mode: "start",
        days: 3,
      })
    ).toEqual({
      startAt: "2026-08-08T09:00:00.000Z",
      endAt: "2026-08-12T17:00:00.000Z",
    });
    expect(
      applyTimelineDrag({
        startAt: "2026-08-05T09:00:00.000Z",
        endAt: "2026-08-12T17:00:00.000Z",
        mode: "end",
        days: -2,
      })
    ).toEqual({
      startAt: "2026-08-05T09:00:00.000Z",
      endAt: "2026-08-10T17:00:00.000Z",
    });
    expect(
      applyTimelineDrag({
        startAt: "2026-08-05T09:00:00.000Z",
        endAt: "2026-08-12T17:00:00.000Z",
        mode: "start",
        days: 20,
      })
    ).toEqual({
      startAt: "2026-08-12T17:00:00.000Z",
      endAt: "2026-08-12T17:00:00.000Z",
    });
    expect(
      applyTimelineDrag({
        startAt: "2026-08-05T09:00:00.000Z",
        endAt: "2026-08-12T17:00:00.000Z",
        mode: "end",
        days: -20,
      })
    ).toEqual({
      startAt: "2026-08-05T09:00:00.000Z",
      endAt: "2026-08-05T09:00:00.000Z",
    });
    expect(
      applyTimelineDrag({
        startAt: "2026-08-05T09:00:00.000Z",
        endAt: "2026-08-12T17:00:00.000Z",
        mode: "move",
        days: 0,
      })
    ).toEqual({
      startAt: "2026-08-05T09:00:00.000Z",
      endAt: "2026-08-12T17:00:00.000Z",
    });
  });

  it("classifies a dragged span as work or a milestone", () => {
    expect(
      timelineRangeFromSpan(
        "2026-08-05T09:00:00.000Z",
        "2026-08-12T17:00:00.000Z"
      )
    ).toEqual({
      startAt: "2026-08-05T09:00:00.000Z",
      endAt: "2026-08-12T17:00:00.000Z",
      kind: "work",
    });
    expect(
      timelineRangeFromSpan(
        "2026-08-12T17:00:00.000Z",
        "2026-08-12T17:00:00.000Z"
      )
    ).toEqual({
      startAt: "2026-08-12T17:00:00.000Z",
      endAt: "2026-08-12T17:00:00.000Z",
      kind: "milestone",
    });
  });

  it("maps a dragged range back onto work start and due dates", () => {
    expect(
      workItemDatesFromTimelineRange(
        {
          startDate: "2026-08-05T09:00:00.000Z",
          dueDate: "2026-08-12T17:00:00.000Z",
        },
        {
          startAt: "2026-08-07T09:00:00.000Z",
          endAt: "2026-08-14T17:00:00.000Z",
        }
      )
    ).toEqual({
      startDate: "2026-08-07T09:00:00.000Z",
      dueDate: "2026-08-14T17:00:00.000Z",
    });
    expect(
      workItemDatesFromTimelineRange(
        { startDate: null, dueDate: "2026-08-12T17:00:00.000Z" },
        {
          startAt: "2026-08-14T17:00:00.000Z",
          endAt: "2026-08-14T17:00:00.000Z",
        }
      )
    ).toEqual({
      startDate: null,
      dueDate: "2026-08-14T17:00:00.000Z",
    });
    expect(
      workItemDatesFromTimelineRange(
        { startDate: "2026-08-05T09:00:00.000Z", dueDate: null },
        {
          startAt: "2026-08-03T09:00:00.000Z",
          endAt: "2026-08-03T09:00:00.000Z",
        }
      )
    ).toEqual({
      startDate: "2026-08-03T09:00:00.000Z",
      dueDate: null,
    });
    expect(
      workItemDatesFromTimelineRange(
        { startDate: null, dueDate: "2026-08-12T17:00:00.000Z" },
        {
          startAt: "2026-08-05T09:00:00.000Z",
          endAt: "2026-08-12T17:00:00.000Z",
        }
      )
    ).toEqual({
      startDate: "2026-08-05T09:00:00.000Z",
      dueDate: "2026-08-12T17:00:00.000Z",
    });
  });
});

function workItem(
  overrides: Pick<WorkItem, "id" | "startDate" | "dueDate">
): WorkItem {
  return {
    dataSource: "api",
    key: "PLAT-1",
    title: "Timeline item",
    summary: "",
    namespaceId: "ns-1",
    projectId: "proj-1",
    status: "backlog",
    priority: "normal",
    assigneeIds: [],
    reviewerIds: [],
    assigneeId: null,
    creatorId: "user-1",
    labelIds: [],
    rank: 1,
    createdAt: "2026-08-01T00:00:00.000Z",
    updatedAt: "2026-08-01T00:00:00.000Z",
    ...overrides,
  };
}
