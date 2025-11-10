import { useMutation } from "@tanstack/react-query";
import { CheckCircle, Edit, Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useEditTodoForm } from "@/contexts/edit-todo-form-context";
import type { TodoPriority } from "@/lib/api";
import {
  v1TodoDeleteMutation,
  v1TodoUpdateMutation,
} from "@/lib/client/@tanstack/react-query.gen";
import { showErrorToast, showSuccessToast } from "@/lib/toast";
import { formatDate } from "@/lib/utils";

interface TodoItemProps {
  todo: {
    id: string;
    title: string;
    description: string;
    priority: TodoPriority;
    completed: boolean;
    due_date: string | null;
    created_at: string;
  };
  getPriorityColor: (priority: TodoPriority) => string;
  onSuccess?: () => void;
}

export function TodoItem({ todo, getPriorityColor, onSuccess }: TodoItemProps) {
  const updateMutation = useMutation(v1TodoUpdateMutation());
  const deleteMutation = useMutation(v1TodoDeleteMutation());
  const { open: openEditForm } = useEditTodoForm();

  const handleToggleComplete = () => {
    updateMutation.mutate(
      {
        path: { id: todo.id },
        body: { completed: !todo.completed },
      },
      {
        onSuccess: () => {
          onSuccess?.();
          showSuccessToast(
            "Todo updated",
            `Todo "${todo.title}" marked as ${!todo.completed ? "completed" : "incomplete"}`
          );
        },
        onError: (error) => {
          showErrorToast("Failed to update todo", error.message);
        },
      }
    );
  };

  const handleDelete = () => {
    deleteMutation.mutate(
      {
        path: { id: todo.id },
      },
      {
        onSuccess: () => {
          onSuccess?.();
          showSuccessToast(
            "Todo deleted",
            `Todo "${todo.title}" has been deleted`
          );
        },
        onError: (error) => {
          showErrorToast("Failed to delete todo", error.message);
        },
      }
    );
  };

  const handleEdit = () => {
    openEditForm(todo);
  };
  return (
    <div
      className={`group bg-background relative rounded-lg border p-4 transition-all hover:shadow-sm ${
        todo.completed ? "opacity-75" : ""
      }`}
    >
      <div className="mb-3 flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-start gap-2">
            <h4
              className={`text-sm leading-tight font-medium ${
                todo.completed ? "text-muted-foreground line-through" : ""
              }`}
            >
              {todo.title}
            </h4>
            <Badge
              variant={getPriorityColor(todo.priority) as any}
              className="shrink-0 rounded px-1.5 py-0.5 text-xs font-medium"
            >
              {todo.priority}
            </Badge>
          </div>
        </div>

        {todo.completed && (
          <Badge
            variant="success"
            className="shrink-0 rounded px-1.5 py-0.5 text-xs"
          >
            Completed
          </Badge>
        )}
      </div>

      {todo.description && (
        <p
          className={`text-muted-foreground mb-3 text-xs leading-relaxed ${
            todo.completed ? "line-through" : ""
          }`}
        >
          {todo.description}
        </p>
      )}

      <div className="flex items-center justify-between">
        <div className="text-muted-foreground text-xs">
          <span>Due: {formatDate(todo.due_date)}</span>
        </div>

        <div className="flex items-center gap-1 opacity-0 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100">
          <Button
            size="sm"
            variant="ghost"
            onClick={handleToggleComplete}
            disabled={updateMutation.isPending || deleteMutation.isPending}
            className={`hover:bg-muted size-7 p-0 ${todo.completed ? "text-green-600" : "text-muted-foreground hover:text-green-600"}`}
            title={todo.completed ? "Mark as incomplete" : "Mark as complete"}
          >
            <CheckCircle className="size-4" />
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={handleEdit}
            disabled={
              todo.completed ||
              updateMutation.isPending ||
              deleteMutation.isPending
            }
            className="hover:bg-muted size-7 p-0"
            title="Edit todo"
          >
            <Edit className="size-4" />
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={handleDelete}
            disabled={deleteMutation.isPending || updateMutation.isPending}
            className="text-destructive hover:bg-destructive/10 hover:text-destructive size-7 p-0"
            title="Delete todo"
          >
            <Trash2 className="size-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
