import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CalendarIcon, Edit, Trash2 } from "lucide-react";

import { StatusIndicator } from "@/components/shared/status-indicator";
import { todoPriorityTone } from "@/components/todo/priority";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  v1TodoDeleteMutation,
  v1TodoUpdateMutation,
} from "@/lib/api/mutation-options";
import { v1TodosGetOptions } from "@/lib/api/query-options";
import type { Todo } from "@/lib/api/types";
import { formatDate } from "@/lib/format-date";
import { rollbackOptimisticQueryData } from "@/lib/mutation-workflow";
import { showErrorToast, showSuccessToast } from "@/lib/toast";
import { uiActions } from "@/lib/ui-store";
import { cn } from "@/lib/utils";

interface TodoItemProps {
  todo: Todo;
  onSuccess?: () => void | Promise<void>;
}

function getDueDateLabel(dueDate: string | null) {
  return dueDate ? formatDate(dueDate) : "No due date";
}

export function TodoItem({ todo, onSuccess }: TodoItemProps) {
  const queryClient = useQueryClient();
  const todosQueryKey = v1TodosGetOptions().queryKey;

  const updateMutation = useMutation({
    ...v1TodoUpdateMutation(),
    onMutate: async (variables) => {
      await queryClient.cancelQueries({ queryKey: todosQueryKey });
      const previous = queryClient.getQueryData<Todo[]>(todosQueryKey);
      const completed = variables.body?.completed;
      const nextTodos = (previous ?? []).map((item) =>
        item.id === variables.path.id && completed !== undefined
          ? { ...item, completed }
          : item
      );
      queryClient.setQueryData<Todo[]>(todosQueryKey, nextTodos);
      return { previous };
    },
    onError: (_error, _variables, context) => {
      rollbackOptimisticQueryData(queryClient, todosQueryKey, context);
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: todosQueryKey }),
  });

  const deleteMutation = useMutation({
    ...v1TodoDeleteMutation(),
    onMutate: async (variables) => {
      await queryClient.cancelQueries({ queryKey: todosQueryKey });
      const previous = queryClient.getQueryData<Todo[]>(todosQueryKey);
      const nextTodos = (previous ?? []).filter(
        (item) => item.id !== variables.path.id
      );
      queryClient.setQueryData<Todo[]>(todosQueryKey, nextTodos);
      return { previous };
    },
    onError: (_error, _variables, context) => {
      rollbackOptimisticQueryData(queryClient, todosQueryKey, context);
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: todosQueryKey }),
  });

  const handleToggleComplete = async () => {
    if (updateMutation.isPending || deleteMutation.isPending) {
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
    if (updateMutation.isPending || deleteMutation.isPending) {
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
        "group/todo hover:bg-muted/50 flex items-start gap-3 px-3 py-2.5 transition-colors",
        todo.completed && "opacity-80"
      )}
    >
      <Checkbox
        checked={todo.completed}
        onCheckedChange={() => {
          void handleToggleComplete();
        }}
        disabled={updateMutation.isPending || deleteMutation.isPending}
        aria-label={completeLabel}
        aria-describedby={titleId}
        className="mt-0.5"
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
          <p
            className={cn(
              "text-muted-foreground mt-0.5 line-clamp-1 text-xs",
              todo.completed && "line-through"
            )}
          >
            {todo.description}
          </p>
        )}

        <div className="mt-1.5 flex flex-wrap items-center gap-2">
          <StatusIndicator
            status={todo.priority}
            tone={todoPriorityTone[todo.priority]}
            className="shrink-0 rounded-sm"
          />
          <span className="text-muted-foreground inline-flex items-center gap-1 text-xs">
            <CalendarIcon className="size-3" />
            {getDueDateLabel(todo.due_date)}
          </span>
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-focus-within/todo:opacity-100 sm:group-hover/todo:opacity-100">
        {!todo.completed && (
          <Button
            size="icon-xs"
            variant="ghost"
            onClick={handleEdit}
            disabled={updateMutation.isPending || deleteMutation.isPending}
            aria-label="Edit todo"
            title="Edit todo"
          >
            <Edit />
          </Button>
        )}
        <Button
          size="icon-xs"
          variant="destructive-ghost"
          onClick={() => {
            void handleDelete();
          }}
          disabled={updateMutation.isPending || deleteMutation.isPending}
          aria-label="Delete todo"
          title="Delete todo"
        >
          <Trash2 />
        </Button>
      </div>
    </div>
  );
}
