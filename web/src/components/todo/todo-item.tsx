import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { QueryClient, QueryKey } from "@tanstack/react-query";
import { CalendarIcon, Edit, Trash2 } from "lucide-react";

import { TodoPriorityRibbon } from "./todo-priority-ribbon";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { MarkdownContent } from "@/components/work/markdown-content";
import {
  v1TodoDeleteMutation,
  v1TodoUpdateMutation,
} from "@/lib/api/mutation-options";
import { v1TodosGetOptions } from "@/lib/api/query-options";
import type { Todo } from "@/lib/api/types";
import { formatDate } from "@/lib/format-date";
import { showErrorToast, showSuccessToast } from "@/lib/toast";
import { uiActions } from "@/lib/ui-store";
import { cn } from "@/lib/utils";

interface TodoItemProps {
  todo: Todo;
  onSuccess?: () => void | Promise<void>;
}

interface TodosQueryPage {
  items?: Todo[] | null;
}

function getDueDateLabel(dueDate: string | null) {
  return dueDate ? formatDate(dueDate) : "No due date";
}

function isTodosQueryPage(value: unknown): value is TodosQueryPage {
  return typeof value === "object" && value !== null && "items" in value;
}

function patchTodoQueries(
  queryClient: QueryClient,
  queryKey: QueryKey,
  updater: (todos: Todo[]) => Todo[]
) {
  queryClient.setQueriesData({ queryKey }, (previous: unknown) => {
    if (Array.isArray(previous)) {
      return updater(previous as Todo[]);
    }
    if (isTodosQueryPage(previous)) {
      return { ...previous, items: updater(previous.items ?? []) };
    }
    return previous;
  });
}

export function TodoItem({ todo, onSuccess }: TodoItemProps) {
  const queryClient = useQueryClient();
  const todosQueryKey = v1TodosGetOptions().queryKey;

  const updateMutation = useMutation({
    ...v1TodoUpdateMutation(),
    onMutate: async (variables) => {
      await queryClient.cancelQueries({ queryKey: todosQueryKey });
      const previous = queryClient.getQueriesData({ queryKey: todosQueryKey });
      const completed = variables.body?.completed;
      patchTodoQueries(queryClient, todosQueryKey, (todos) =>
        todos.map((item) =>
          item.id === variables.path.id && completed !== undefined
            ? { ...item, completed }
            : item
        )
      );
      return { previous };
    },
    onError: (_error, _variables, context) => {
      context?.previous.forEach(([queryKey, data]) => {
        queryClient.setQueryData(queryKey, data);
      });
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: todosQueryKey }),
  });

  const deleteMutation = useMutation({
    ...v1TodoDeleteMutation(),
    onMutate: async (variables) => {
      await queryClient.cancelQueries({ queryKey: todosQueryKey });
      const previous = queryClient.getQueriesData({ queryKey: todosQueryKey });
      patchTodoQueries(queryClient, todosQueryKey, (todos) =>
        todos.filter((item) => item.id !== variables.path.id)
      );
      return { previous };
    },
    onError: (_error, _variables, context) => {
      context?.previous.forEach(([queryKey, data]) => {
        queryClient.setQueryData(queryKey, data);
      });
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: todosQueryKey }),
  });

  const isBusy = updateMutation.isPending || deleteMutation.isPending;

  const handleToggleComplete = async () => {
    if (isBusy) {
      return;
    }

    const nextCompleted = !todo.completed;

    try {
      await updateMutation.mutateAsync({
        path: { id: todo.id },
        body: { completed: nextCompleted },
      });
      await onSuccess?.();
      showSuccessToast(
        "Todo updated",
        `Todo "${todo.title}" marked as ${nextCompleted ? "completed" : "incomplete"}`
      );
    } catch (error) {
      showErrorToast(
        "Failed to update todo",
        error instanceof Error ? error.message : "Unknown error occurred"
      );
    }
  };

  const handleDelete = async () => {
    if (isBusy) {
      return;
    }

    try {
      await deleteMutation.mutateAsync({
        path: { id: todo.id },
      });
      await onSuccess?.();
      showSuccessToast("Todo deleted", `Todo "${todo.title}" has been deleted`);
    } catch (error) {
      showErrorToast(
        "Failed to delete todo",
        error instanceof Error ? error.message : "Unknown error occurred"
      );
    }
  };

  const handleEdit = () => {
    uiActions.openEditTodo(todo);
  };

  const titleId = `todo-title-${todo.id}`;
  const completeLabel = todo.completed
    ? `Mark "${todo.title}" as incomplete`
    : `Mark "${todo.title}" as complete`;

  return (
    <div
      role="listitem"
      className={cn(
        "group/todo hover:bg-muted/50 flex min-w-0 items-start gap-3 px-3 py-2.5 transition-colors",
        todo.completed && "opacity-80"
      )}
    >
      <Checkbox
        checked={todo.completed}
        disabled={isBusy}
        aria-label={completeLabel}
        aria-describedby={titleId}
        className="mt-0.5"
        onClick={() => {
          void handleToggleComplete();
        }}
      />

      <div className="min-w-0 flex-1">
        <p
          id={titleId}
          className={cn(
            "text-sm font-medium",
            todo.completed && "text-muted-foreground line-through"
          )}
        >
          {todo.title}
        </p>

        {todo.description && (
          <MarkdownContent
            markdown={todo.description}
            size="xs"
            className={cn(
              "mt-0.5 line-clamp-1",
              todo.completed && "line-through"
            )}
          />
        )}

        <div className="mt-1.5 flex flex-wrap items-center gap-2">
          <TodoPriorityRibbon
            priority={todo.priority}
            className="gap-1"
            iconClassName="size-3"
            labelClassName="text-xs font-medium"
          />
          <span className="text-muted-foreground inline-flex items-center gap-1 text-xs">
            <CalendarIcon className="size-3" />
            {getDueDateLabel(todo.due_date)}
          </span>
        </div>
      </div>

      <div className="relative z-10 flex shrink-0 items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-focus-within/todo:opacity-100 sm:group-hover/todo:opacity-100">
        {!todo.completed && (
          <Button
            type="button"
            size="icon-xs"
            variant="ghost"
            onClick={handleEdit}
            disabled={isBusy}
            aria-label="Edit todo"
            title="Edit todo"
          >
            <Edit />
          </Button>
        )}
        <Button
          type="button"
          size="icon-xs"
          variant="destructive-ghost"
          onClick={() => {
            void handleDelete();
          }}
          disabled={isBusy}
          aria-label="Delete todo"
          title="Delete todo"
        >
          <Trash2 />
        </Button>
      </div>
    </div>
  );
}
