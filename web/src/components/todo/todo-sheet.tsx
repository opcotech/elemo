"use client";

import { CircleCheckBig, Plus } from "lucide-react";
import { useMemo } from "react";

import { TodoItem } from "@/components/todo";
import { AddTodoForm } from "@/components/todo/add-todo-form";
import { EditTodoForm } from "@/components/todo/edit-todo-form";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import { useAddTodoForm } from "@/contexts/add-todo-form-context";
import { useEditTodoForm } from "@/contexts/edit-todo-form-context";
import { useTodoSheet } from "@/contexts/todo-sheet-context";
import { useTodos } from "@/hooks/use-todos";
import type { TodoPriority } from "@/lib/api";

export function TodoSheetTrigger() {
  const { open } = useTodoSheet();
  const { data: todos } = useTodos();
  const uncompletedCount = todos?.filter((t) => !t.completed).length || 0;

  return (
    <Button
      variant="link"
      size="icon"
      className="text-foreground hover:text-primary relative"
      onClick={open}
      aria-label="Show todo list"
    >
      <CircleCheckBig className="h-5 w-5" />
      {uncompletedCount > 0 && (
        <Badge
          className="absolute top-1 right-1 size-2 rounded-full border-red-600 bg-red-500 p-0 dark:bg-red-500"
          variant="destructive"
        />
      )}
    </Button>
  );
}

export function TodoSheet() {
  const { data: todos, isLoading, refetch } = useTodos();
  const { isOpen, close } = useTodoSheet();
  const {
    isOpen: isAddFormOpen,
    open: openAddForm,
    close: closeAddForm,
  } = useAddTodoForm();
  const {
    isOpen: isEditFormOpen,
    todo: editTodo,
    close: closeEditForm,
  } = useEditTodoForm();

  // Sort todos by due date, priority, then creation date
  const sortedTodos = useMemo(() => {
    if (!todos) return [];

    return [...todos].sort((a, b) => {
      // First sort by completed status (completed todos go to the end)
      if (a.completed && !b.completed) return 1;
      if (!a.completed && b.completed) return -1;

      // Then sort by due date (null dates go to the end)
      const aDueDate = a.due_date ? new Date(a.due_date) : null;
      const bDueDate = b.due_date ? new Date(b.due_date) : null;

      if (aDueDate && !bDueDate) return -1;
      if (!aDueDate && bDueDate) return 1;
      if (aDueDate && bDueDate) {
        const dueDateDiff = aDueDate.getTime() - bDueDate.getTime();
        if (dueDateDiff !== 0) return dueDateDiff;
      }

      // Then sort by priority (critical > urgent > important > normal)
      const priorityOrder = { critical: 4, urgent: 3, important: 2, normal: 1 };
      const aPriority =
        priorityOrder[a.priority as keyof typeof priorityOrder] || 1;
      const bPriority =
        priorityOrder[b.priority as keyof typeof priorityOrder] || 1;

      if (aPriority !== bPriority) return bPriority - aPriority;

      // Finally sort by creation date (newest first)
      return (
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      );
    });
  }, [todos]);

  const formatDate = (dateString: string | null) => {
    if (!dateString) return "No due date";
    return new Date(dateString).toLocaleDateString();
  };

  const getPriorityColor = (priority: TodoPriority) => {
    switch (priority) {
      case "critical":
        return "destructive";
      case "urgent":
        return "warning";
      case "important":
        return "default";
      default:
        return "secondary";
    }
  };

  return (
    <Sheet
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) close();
      }}
    >
      <SheetContent className="sm:w-md sm:max-w-full">
        <SheetHeader className="pb-4">
          <SheetTitle>Todo Items</SheetTitle>
        </SheetHeader>
        <Button
          variant="outline"
          size="sm"
          onClick={openAddForm}
          className="h-8 px-3"
        >
          <Plus className="mr-1.5 h-4 w-4" />
          Add Todo
        </Button>
        <ScrollArea className="h-full">
          {isLoading ? (
            <div className="space-y-3">
              <Skeleton className="h-24 w-full" />
              <Skeleton className="h-24 w-full" />
              <Skeleton className="h-24 w-full" />
            </div>
          ) : sortedTodos.length === 0 ? (
            <div className="flex h-32 items-center justify-center">
              <p className="text-muted-foreground">No todos found</p>
            </div>
          ) : (
            <div className="space-y-4 pr-2">
              {sortedTodos.map((todo) => (
                <TodoItem
                  key={todo.id}
                  todo={todo}
                  formatDate={formatDate}
                  getPriorityColor={getPriorityColor}
                  onSuccess={() => {
                    refetch();
                  }}
                />
              ))}
            </div>
          )}
        </ScrollArea>
      </SheetContent>

      <AddTodoForm
        open={isAddFormOpen}
        onOpenChange={closeAddForm}
        onSuccess={() => {
          refetch();
        }}
      />

      <EditTodoForm
        open={isEditFormOpen}
        onOpenChange={closeEditForm}
        onSuccess={() => {
          refetch();
        }}
        todo={editTodo}
      />
    </Sheet>
  );
}
