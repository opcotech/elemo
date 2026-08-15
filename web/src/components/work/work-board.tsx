import {
  DndContext,
  DragOverlay,
  PointerSensor,
  closestCorners,
  pointerWithin,
  useDroppable,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import type {
  CollisionDetection,
  DragEndEvent,
  DragOverEvent,
  DragStartEvent,
  UniqueIdentifier,
} from "@dnd-kit/core";
import {
  SortableContext,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { PlusIcon } from "lucide-react";
import { useEffect, useId, useMemo, useState } from "react";

import { issuePriorityLabels } from "./priority-ribbon";
import type { BoardItemMove, BoardMoveGroup } from "./use-board-issue-move";
import { WorkCard } from "./work-card";

import { openQuickCreate } from "@/components/quick-create/open";
import { Button } from "@/components/ui/button";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { getPerson } from "@/lib/mock-data";
import type { WorkItem, WorkPriority, WorkStatus } from "@/lib/mock-data";
import { cn } from "@/lib/utils";
import type { WorkRouteSearch } from "@/lib/work-route-search";

const COLUMN_PAGE_SIZE = 25;
const statusOrder: readonly WorkStatus[] = [
  "backlog",
  "in progress",
  "in review",
  "blocked",
  "done",
  "closed",
];
const priorityOrder: readonly WorkPriority[] = [
  "highest",
  "high",
  "normal",
  "low",
  "lowest",
];

function columnId(key: string) {
  return `column:${key}`;
}

function parseColumnId(id: UniqueIdentifier): string | null {
  const value = String(id);
  return value.startsWith("column:") ? value.slice("column:".length) : null;
}

const boardCollisionDetection: CollisionDetection = (args) => {
  const pointerHits = pointerWithin(args);
  if (pointerHits.length > 0) {
    return pointerHits;
  }
  return closestCorners(args);
};

function isBoardMoveGroup(
  group: WorkRouteSearch["group"]
): group is BoardMoveGroup {
  return group === "status" || group === "priority";
}

function DraggableWorkCard({
  item,
  compact,
  onSelect,
  disabled,
}: {
  item: WorkItem;
  compact: boolean;
  onSelect: (item: WorkItem) => void;
  disabled: boolean;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: item.id,
    disabled,
    data: { type: "item", item },
  });

  return (
    <div
      ref={setNodeRef}
      style={{
        transform: CSS.Transform.toString(transform),
        transition,
      }}
      className={cn(isDragging && "opacity-40")}
      {...attributes}
      {...listeners}
    >
      <WorkCard
        item={item}
        compact={compact}
        onSelect={onSelect}
        className={disabled ? undefined : "cursor-grab active:cursor-grabbing"}
      />
    </div>
  );
}

function BoardColumn({
  columnKey,
  label,
  items,
  compact,
  onSelect,
  dragEnabled,
}: {
  columnKey: string;
  label: string;
  items: readonly WorkItem[];
  compact: boolean;
  onSelect: (item: WorkItem) => void;
  dragEnabled: boolean;
}) {
  const headingId = useId();
  const [visibleCount, setVisibleCount] = useState(COLUMN_PAGE_SIZE);
  const visibleItems = items.slice(0, visibleCount);
  const remaining = items.length - visibleItems.length;
  const { setNodeRef, isOver } = useDroppable({
    id: columnId(columnKey),
    data: { type: "column", columnKey },
    disabled: !dragEnabled,
  });

  return (
    <section
      aria-labelledby={headingId}
      className={cn(
        "bg-surface-sunken flex max-h-[calc(100svh-15rem)] w-72 flex-col rounded-xl border",
        isOver && "ring-primary/40 ring-2"
      )}
    >
      <header className="flex shrink-0 items-center gap-2 border-b px-3 py-2.5">
        <h2
          id={headingId}
          className="flex-1 text-xs font-semibold tracking-wide uppercase"
        >
          {label.replaceAll("-", " ")}
        </h2>
        <span className="text-muted-foreground text-xs tabular-nums">
          {items.length}
        </span>
        <Button
          variant="ghost"
          size="icon-xs"
          onClick={() => openQuickCreate("work")}
          aria-label={`Add work to ${label}`}
        >
          <PlusIcon />
        </Button>
      </header>
      <div ref={setNodeRef} className="min-h-20 space-y-2 overflow-y-auto p-2">
        <SortableContext
          id={columnKey}
          items={visibleItems.map((item) => item.id)}
          strategy={verticalListSortingStrategy}
        >
          {visibleItems.map((item) => (
            <DraggableWorkCard
              key={item.id}
              item={item}
              compact={compact}
              onSelect={onSelect}
              disabled={!dragEnabled}
            />
          ))}
        </SortableContext>
        {items.length === 0 && (
          <p className="text-muted-foreground py-6 text-center text-xs">
            No work in this group
          </p>
        )}
        {remaining > 0 && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="w-full"
            onClick={() => setVisibleCount((count) => count + COLUMN_PAGE_SIZE)}
          >
            Show {Math.min(remaining, COLUMN_PAGE_SIZE)} more
          </Button>
        )}
      </div>
    </section>
  );
}

function findColumnForItem(
  columns: ReadonlyMap<string, readonly WorkItem[]>,
  itemId: UniqueIdentifier
): string | null {
  for (const [key, columnItems] of columns) {
    if (columnItems.some((item) => item.id === itemId)) {
      return key;
    }
  }
  return null;
}

export function resolveBoardDropColumn(
  columns: ReadonlyMap<string, readonly WorkItem[]>,
  activeId: UniqueIdentifier,
  overId: UniqueIdentifier
): string | null {
  return (
    parseColumnId(overId) ??
    findColumnForItem(columns, overId) ??
    findColumnForItem(columns, activeId)
  );
}

export function WorkBoard({
  items,
  group,
  compact,
  onSelect,
  onItemMove,
}: {
  items: readonly WorkItem[];
  group: WorkRouteSearch["group"];
  compact: boolean;
  onSelect: (item: WorkItem) => void;
  onItemMove?: (move: BoardItemMove) => void;
}) {
  const dragEnabled = isBoardMoveGroup(group);
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 8 },
    })
  );

  const groupedFromProps = useMemo(() => {
    const groupValue = (item: WorkItem) => {
      if (group === "priority") return item.priority;
      if (group === "assignee") {
        const assignee = item.assignees?.[0];
        if (assignee) {
          return assignee.name;
        }
        return item.assigneeId
          ? (getPerson(item.assigneeId)?.displayName ?? "Assigned")
          : "Unassigned";
      }
      if (group === "none") return "All work";
      return item.status;
    };
    const grouped = Map.groupBy(items, groupValue);
    const keys =
      group === "status"
        ? statusOrder
        : group === "priority"
          ? priorityOrder
          : ([...grouped.keys()] as readonly string[]);

    return {
      keys,
      columns: new Map(
        keys.map((key) => [key, [...(grouped.get(key) ?? [])]] as const)
      ),
    };
  }, [group, items]);

  const [columns, setColumns] = useState(groupedFromProps.columns);
  const [activeId, setActiveId] = useState<UniqueIdentifier | null>(null);

  useEffect(() => {
    setColumns(groupedFromProps.columns);
  }, [groupedFromProps.columns]);

  const activeItem = useMemo(() => {
    if (!activeId) return undefined;
    for (const columnItems of columns.values()) {
      const match = columnItems.find((item) => item.id === activeId);
      if (match) return match;
    }
    return items.find((item) => item.id === activeId);
  }, [activeId, columns, items]);

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id);
  };

  const handleDragOver = (event: DragOverEvent) => {
    if (!dragEnabled) return;

    const { active, over } = event;
    if (!over) return;

    const activeColumn = findColumnForItem(columns, active.id);
    const overColumn =
      parseColumnId(over.id) ?? findColumnForItem(columns, over.id);
    if (!activeColumn || !overColumn || activeColumn === overColumn) {
      return;
    }

    setColumns((previous) => {
      const sourceItems = previous.get(activeColumn);
      const targetItems = previous.get(overColumn);
      if (!sourceItems || !targetItems) return previous;

      const activeIndex = sourceItems.findIndex(
        (item) => item.id === active.id
      );
      if (activeIndex < 0) return previous;

      const moving = sourceItems[activeIndex];
      if (!moving) return previous;

      const nextSource = sourceItems.filter((item) => item.id !== active.id);
      const overIndex = targetItems.findIndex((item) => item.id === over.id);
      const insertAt = overIndex >= 0 ? overIndex : targetItems.length;
      const nextTarget = [...targetItems];
      nextTarget.splice(insertAt, 0, {
        ...moving,
        ...(group === "status"
          ? { status: overColumn as WorkStatus }
          : { priority: overColumn as WorkPriority }),
      });

      const next = new Map(previous);
      next.set(activeColumn, nextSource);
      next.set(overColumn, nextTarget);
      return next;
    });
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveId(null);
    if (!dragEnabled || !over || !onItemMove) {
      setColumns(groupedFromProps.columns);
      return;
    }

    const from = findColumnForItem(groupedFromProps.columns, active.id);
    const to = resolveBoardDropColumn(columns, active.id, over.id);

    const item = items.find((candidate) => candidate.id === active.id);

    if (!from || !to || !item || from === to) {
      setColumns(groupedFromProps.columns);
      return;
    }

    onItemMove({
      item,
      group,
      from,
      to,
    });
  };

  const handleDragCancel = () => {
    setActiveId(null);
    setColumns(groupedFromProps.columns);
  };

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={boardCollisionDetection}
      onDragStart={handleDragStart}
      onDragOver={handleDragOver}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
    >
      <ScrollArea className="h-full min-h-0 min-w-0">
        <div className="flex min-w-max items-start gap-3">
          {groupedFromProps.keys.map((key) => (
            <BoardColumn
              key={key}
              columnKey={key}
              label={
                group === "priority"
                  ? issuePriorityLabels[key as WorkPriority]
                  : key
              }
              items={columns.get(key) ?? []}
              compact={compact}
              onSelect={onSelect}
              dragEnabled={dragEnabled}
            />
          ))}
        </div>
        <ScrollBar orientation="horizontal" />
      </ScrollArea>
      <DragOverlay dropAnimation={null}>
        {activeItem ? (
          <WorkCard
            item={activeItem}
            compact={compact}
            className="cursor-grabbing shadow-lg"
          />
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
