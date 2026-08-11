import { describe, expect, it } from "vitest";

import { getTodoDueGroup, groupTodosByDueDate } from "./grouping";

import type { Todo } from "@/lib/api/types";

/** Wednesday 2026-08-12 — week Mon 10 – Sun 16 */
const NOW = new Date("2026-08-12T15:30:00.000Z");

function makeTodo(overrides: Partial<Todo> & Pick<Todo, "id" | "title">): Todo {
  return {
    description: "",
    priority: "normal",
    completed: false,
    due_date: null,
    owned_by: "user-1",
    created_by: "user-1",
    created_at: "2026-08-12T10:00:00.000Z",
    updated_at: "2026-08-12T10:00:00.000Z",
    ...overrides,
  };
}

describe("getTodoDueGroup", () => {
  it("puts null and missing due dates in later", () => {
    expect(getTodoDueGroup(null, NOW)).toBe("later");
    expect(getTodoDueGroup(undefined, NOW)).toBe("later");
  });

  it("puts today and overdue in today", () => {
    expect(getTodoDueGroup("2026-08-12T08:00:00.000Z", NOW)).toBe("today");
    expect(getTodoDueGroup("2026-08-11T23:59:00.000Z", NOW)).toBe("today");
    expect(getTodoDueGroup("2026-08-01T12:00:00.000Z", NOW)).toBe("today");
  });

  it("puts tomorrow in tomorrow", () => {
    expect(getTodoDueGroup("2026-08-13T09:00:00.000Z", NOW)).toBe("tomorrow");
  });

  it("puts remaining weekdays in this_week", () => {
    expect(getTodoDueGroup("2026-08-14T09:00:00.000Z", NOW)).toBe("this_week");
    expect(getTodoDueGroup("2026-08-16T09:00:00.000Z", NOW)).toBe("this_week");
  });

  it("puts dates after this week in later", () => {
    expect(getTodoDueGroup("2026-08-17T09:00:00.000Z", NOW)).toBe("later");
    expect(getTodoDueGroup("2026-09-01T09:00:00.000Z", NOW)).toBe("later");
  });
});

describe("groupTodosByDueDate", () => {
  it("returns only non-empty groups in order", () => {
    const todos = [
      makeTodo({ id: "1", title: "No date", due_date: null }),
      makeTodo({
        id: "2",
        title: "Today",
        due_date: "2026-08-12T12:00:00.000Z",
      }),
      makeTodo({
        id: "3",
        title: "Friday",
        due_date: "2026-08-14T12:00:00.000Z",
      }),
    ];

    const groups = groupTodosByDueDate(todos, NOW);

    expect(groups.map((group) => group.id)).toEqual([
      "today",
      "this_week",
      "later",
    ]);
    expect(groups.map((group) => group.label)).toEqual([
      "Today",
      "This week",
      "Later",
    ]);
  });

  it("sorts within a group by completed, due date, priority, then created", () => {
    const todos = [
      makeTodo({
        id: "low",
        title: "Low",
        due_date: "2026-08-12T12:00:00.000Z",
        priority: "normal",
        created_at: "2026-08-12T09:00:00.000Z",
      }),
      makeTodo({
        id: "done",
        title: "Done",
        due_date: "2026-08-12T08:00:00.000Z",
        completed: true,
        priority: "critical",
      }),
      makeTodo({
        id: "high",
        title: "High",
        due_date: "2026-08-12T12:00:00.000Z",
        priority: "urgent",
        created_at: "2026-08-12T08:00:00.000Z",
      }),
    ];

    const [today] = groupTodosByDueDate(todos, NOW);

    expect(today.todos.map((todo) => todo.id)).toEqual(["high", "low", "done"]);
  });
});
