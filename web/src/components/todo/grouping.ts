import { addDays, endOfWeek, isBefore, isSameDay, startOfDay } from "date-fns";

import type { Todo, TodoPriority } from "@/lib/api/types";

export type TodoDueGroupId = "today" | "tomorrow" | "this_week" | "later";

export interface TodoDueGroup {
  id: TodoDueGroupId;
  label: string;
  todos: Todo[];
}

const GROUP_ORDER: readonly TodoDueGroupId[] = [
  "today",
  "tomorrow",
  "this_week",
  "later",
] as const;

const GROUP_LABELS: Record<TodoDueGroupId, string> = {
  today: "Today",
  tomorrow: "Tomorrow",
  this_week: "This week",
  later: "Later",
};

const PRIORITY_ORDER: Record<TodoPriority, number> = {
  critical: 4,
  urgent: 3,
  important: 2,
  normal: 1,
};

export function getTodoDueGroup(
  dueDate: string | null | undefined,
  now: Date = new Date()
): TodoDueGroupId {
  if (!dueDate) {
    return "later";
  }

  const due = startOfDay(new Date(dueDate));
  const today = startOfDay(now);
  const tomorrow = addDays(today, 1);
  const weekEnd = startOfDay(endOfWeek(today, { weekStartsOn: 1 }));

  if (isSameDay(due, today) || isBefore(due, today)) {
    return "today";
  }
  if (isSameDay(due, tomorrow)) {
    return "tomorrow";
  }
  if (isBefore(due, addDays(weekEnd, 1))) {
    return "this_week";
  }
  return "later";
}

function compareTodos(a: Todo, b: Todo): number {
  if (a.completed && !b.completed) return 1;
  if (!a.completed && b.completed) return -1;

  const aDueDate = a.due_date ? new Date(a.due_date) : null;
  const bDueDate = b.due_date ? new Date(b.due_date) : null;

  if (aDueDate && !bDueDate) return -1;
  if (!aDueDate && bDueDate) return 1;
  if (aDueDate && bDueDate) {
    const dueDateDiff = aDueDate.getTime() - bDueDate.getTime();
    if (dueDateDiff !== 0) return dueDateDiff;
  }

  const aPriority = PRIORITY_ORDER[a.priority] ?? 1;
  const bPriority = PRIORITY_ORDER[b.priority] ?? 1;
  if (aPriority !== bPriority) return bPriority - aPriority;

  return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
}

export function groupTodosByDueDate(
  todos: Todo[],
  now: Date = new Date()
): TodoDueGroup[] {
  const buckets: Record<TodoDueGroupId, Todo[]> = {
    today: [],
    tomorrow: [],
    this_week: [],
    later: [],
  };

  for (const todo of todos) {
    buckets[getTodoDueGroup(todo.due_date, now)].push(todo);
  }

  return GROUP_ORDER.flatMap((id) => {
    const groupTodos = buckets[id];
    if (groupTodos.length === 0) {
      return [];
    }
    return [
      {
        id,
        label: GROUP_LABELS[id],
        todos: [...groupTodos].sort(compareTodos),
      },
    ];
  });
}
