"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { CircleCheckBig, Plus } from "lucide-react";
import { useMemo } from "react";

import { TodoItem } from "./todo-item";

import {
  CursorPaginator,
  cursorPaginatorProps,
} from "@/components/shared/cursor-paginator";
import { AppList } from "@/components/shared/entity-link";
import { AddTodoForm } from "@/components/todo/add-todo-form";
import { EditTodoForm } from "@/components/todo/edit-todo-form";
import { groupTodosByDueDate } from "@/components/todo/grouping";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import { useCursorPageNav } from "@/hooks/use-cursor-page-nav";
import { cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1TodosGetOptions } from "@/lib/api/query-options";
import { uiActions, useUiSelector } from "@/lib/ui-store";

export function TodoSheet() {
  const pageNav = useCursorPageNav();
  const { data: todosPage, isLoading } = useQuery(
    v1TodosGetOptions({
      query: cursorPageQuery(pageNav.pageToken),
    })
  );
  const todos = todosPage?.items ?? [];
  const queryClient = useQueryClient();
  const isOpen = useUiSelector((state) => state.todoSheetOpen);
  const isAddFormOpen = useUiSelector((state) => state.addTodoOpen);
  const editTodo = useUiSelector((state) => state.editingTodo);
  const invalidateTodos = () =>
    queryClient.invalidateQueries({
      queryKey: v1TodosGetOptions().queryKey,
    });

  const groups = useMemo(() => groupTodosByDueDate(todos), [todos]);
  const openTodoCount = todos.filter((todo) => !todo.completed).length;
  const completedTodoCount = todos.length - openTodoCount;

  return (
    <Sheet
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) uiActions.closeTodoSheet();
      }}
    >
      <SheetContent
        className="gap-4 px-4 pb-4 data-[side=right]:w-full data-[side=right]:sm:max-w-lg"
        data-section="todo-sheet"
      >
        <SheetHeader className="px-0 pt-4 pb-0">
          <SheetTitle>Todo Items</SheetTitle>
          {!isLoading && todos.length > 0 && (
            <p className="text-muted-foreground text-xs">
              {openTodoCount} open · {completedTodoCount} completed
            </p>
          )}
        </SheetHeader>
        <Button
          variant="outline"
          size="sm"
          onClick={uiActions.openAddTodo}
          className="shrink-0"
        >
          <Plus />
          Add Todo
        </Button>
        <ScrollArea className="min-h-0 flex-1">
          {isLoading ? (
            <div className="pr-3">
              <AppList aria-label="Loading todos" aria-busy="true">
                {Array.from({ length: 3 }).map((_, index) => (
                  <div key={index} role="listitem" className="px-3 py-2.5">
                    <Skeleton className="h-10 w-full" />
                  </div>
                ))}
              </AppList>
            </div>
          ) : todos.length === 0 ? (
            <EmptyState
              icon={<CircleCheckBig />}
              title="No todos found"
              description="Create your first todo to get started"
              action={
                <Button
                  onClick={uiActions.openAddTodo}
                  variant="outline"
                  size="sm"
                >
                  <Plus className="size-4" />
                  Add Todo
                </Button>
              }
            />
          ) : (
            <div className="flex flex-col gap-5 pr-3">
              {groups.map((group) => (
                <section key={group.id} className="min-w-0">
                  <h3 className="text-muted-foreground mb-2 text-xs font-semibold tracking-wide uppercase">
                    {group.label}
                  </h3>
                  <AppList
                    aria-label={`${group.label} todos`}
                    className="overflow-visible"
                  >
                    {group.todos.map((todo) => (
                      <TodoItem key={todo.id} todo={todo} />
                    ))}
                  </AppList>
                </section>
              ))}
            </div>
          )}
        </ScrollArea>
        <CursorPaginator {...cursorPaginatorProps(todosPage, pageNav)} />
      </SheetContent>

      <AddTodoForm
        open={isAddFormOpen}
        onOpenChange={(open) => {
          if (!open) uiActions.closeAddTodo();
        }}
        onSuccess={invalidateTodos}
      />

      <EditTodoForm
        open={editTodo !== null}
        onOpenChange={(open) => {
          if (!open) uiActions.closeEditTodo();
        }}
        onSuccess={invalidateTodos}
        todo={editTodo}
      />
    </Sheet>
  );
}
