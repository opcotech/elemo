"use client";

import { useQuery } from "@tanstack/react-query";
import { CircleCheckBig } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { v1TodosGetOptions } from "@/lib/api/query-options";
import { uiActions } from "@/lib/ui-store";

export function TodoSheetTrigger() {
  const { data: todosPage } = useQuery({
    ...v1TodosGetOptions(),
  });
  const uncompletedCount =
    todosPage?.items?.filter((t) => !t.completed).length || 0;

  return (
    <Button
      variant="ghost"
      size="icon"
      className="relative"
      onClick={uiActions.openTodoSheet}
      aria-label="Show todo list"
    >
      <CircleCheckBig />
      {uncompletedCount > 0 && (
        <Badge
          className="border-background bg-destructive absolute top-1 right-1 size-2 rounded-full border p-0"
          variant="destructive"
        />
      )}
    </Button>
  );
}
