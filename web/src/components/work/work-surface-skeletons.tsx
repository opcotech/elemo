import type { ReactNode } from "react";

import { ContentWidth } from "@/components/layout/content-width";
import { Skeleton } from "@/components/ui/skeleton";
import { TableSkeleton } from "@/components/ui/table-skeleton";
import type { TableSkeletonColumn } from "@/components/ui/table-skeleton";
import type { WorkRouteSearch } from "@/lib/work-route-search";

const BOARD_COLUMN_COUNT = 6;
const BOARD_CARDS_PER_COLUMN = 3;
const LIST_ROW_COUNT = 8;
const TIMELINE_ROW_COUNT = 6;
const TIMELINE_LABEL_COLUMN_PX = 300;

const workTableSkeletonColumns: readonly TableSkeletonColumn[] = [
  {
    header: "Key",
    skeletonClassName: "h-3 w-16",
    headerClassName: "min-w-20",
  },
  {
    header: "Title",
    skeletonClassName: "h-4 w-56",
    headerClassName: "min-w-72",
  },
  { header: "Status", skeletonClassName: "h-4 w-20" },
  { header: "Priority", skeletonClassName: "h-4 w-16" },
  { header: "People", skeletonClassName: "size-6 rounded-full" },
  { header: "Labels", skeletonClassName: "h-5 w-16 rounded-full" },
  { header: "Start date", skeletonClassName: "h-3 w-20" },
  { header: "Due date", skeletonClassName: "h-3 w-20" },
];

function LoadingStatus({
  label,
  className,
  children,
}: {
  label: string;
  className?: string;
  children: ReactNode;
}) {
  return (
    <div role="status" aria-busy="true" className={className}>
      <span className="sr-only">Loading {label}</span>
      {children}
    </div>
  );
}

function WorkCardSkeleton() {
  return (
    <article className="bg-card w-full rounded-lg border p-3 shadow-xs">
      <Skeleton className="h-3 w-16" />
      <Skeleton className="mt-1 h-4 w-full" />
      <Skeleton className="mt-2 h-3 w-24" />
      <div className="mt-4 flex items-center gap-3">
        <Skeleton className="size-6 rounded-full" />
        <Skeleton className="h-3 w-16" />
      </div>
    </article>
  );
}

export function WorkBoardSkeleton() {
  return (
    <LoadingStatus
      label="board"
      className="h-full min-h-0 min-w-0 overflow-auto"
    >
      <div className="flex min-w-max items-start gap-3">
        {Array.from({ length: BOARD_COLUMN_COUNT }, (_column, columnIndex) => (
          <section
            key={columnIndex}
            className="bg-surface-sunken flex max-h-[calc(100svh-15rem)] w-72 min-w-72 flex-col rounded-xl border"
          >
            <header className="flex shrink-0 items-center gap-2 border-b px-3 py-2.5">
              <Skeleton className="h-3 w-20" />
              <Skeleton className="ml-auto h-3 w-4" />
            </header>
            <div className="min-h-20 space-y-2 p-2">
              {Array.from(
                { length: BOARD_CARDS_PER_COLUMN },
                (_card, cardIndex) => (
                  <WorkCardSkeleton key={cardIndex} />
                )
              )}
            </div>
          </section>
        ))}
      </div>
    </LoadingStatus>
  );
}

export function WorkListSkeleton() {
  return (
    <LoadingStatus
      label="list"
      className="h-full min-w-0 overflow-auto rounded-lg border"
    >
      {Array.from({ length: LIST_ROW_COUNT }, (_, index) => (
        <div
          key={index}
          className="flex min-w-0 items-center gap-3 border-b px-3 py-2.5 last:border-b-0"
        >
          <Skeleton className="size-2 shrink-0 rounded-full" />
          <Skeleton className="h-3 w-20 shrink-0" />
          <Skeleton className="h-4 min-w-0 flex-1" />
          <Skeleton className="size-6 shrink-0 rounded-full" />
        </div>
      ))}
    </LoadingStatus>
  );
}

export function WorkTableSkeleton() {
  return (
    <LoadingStatus label="table" className="h-full min-h-0 overflow-auto">
      <TableSkeleton columns={workTableSkeletonColumns} rows={8} />
    </LoadingStatus>
  );
}

export function WorkTimelineSkeleton() {
  return (
    <LoadingStatus label="timeline" className="h-full min-h-0 overflow-auto">
      <div
        className="grid border-b"
        style={{
          gridTemplateColumns: `${TIMELINE_LABEL_COLUMN_PX}px minmax(0, 1fr)`,
        }}
      >
        <div className="border-r px-4 py-3">
          <Skeleton className="h-3 w-16" />
        </div>
        <div className="flex items-center gap-8 px-4 py-3">
          <Skeleton className="h-3 w-16" />
          <Skeleton className="h-3 w-16" />
          <Skeleton className="h-3 w-16" />
        </div>
      </div>
      {Array.from({ length: TIMELINE_ROW_COUNT }, (_, index) => (
        <div
          key={index}
          className="grid border-b"
          style={{
            gridTemplateColumns: `${TIMELINE_LABEL_COLUMN_PX}px minmax(0, 1fr)`,
          }}
        >
          <div className="border-r px-4 py-3">
            <Skeleton className="h-4 w-48" />
          </div>
          <div className="flex items-center px-8 py-3">
            <Skeleton
              className="h-6 rounded-full"
              style={{
                width: `${40 + ((index * 17) % 40)}%`,
                marginLeft: `${(index * 11) % 24}%`,
              }}
            />
          </div>
        </div>
      ))}
    </LoadingStatus>
  );
}

export function WorkInspectorSkeleton() {
  return (
    <LoadingStatus label="inspector" className="space-y-6 p-4 pt-4 pr-12">
      <div className="space-y-2">
        <Skeleton className="h-3 w-16" />
        <Skeleton className="h-7 w-56" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-3/4" />
      </div>
      <Skeleton className="h-9 w-full rounded-md" />
      <div className="divide-border/60 divide-y">
        {Array.from({ length: 6 }, (_, index) => (
          <div
            key={index}
            className="grid grid-cols-[minmax(7rem,0.75fr)_minmax(0,1.25fr)] items-start gap-4 py-2"
          >
            <Skeleton className="h-3 w-16" />
            <Skeleton className="h-4 w-28" />
          </div>
        ))}
      </div>
      <div className="space-y-3">
        <Skeleton className="h-3 w-20" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-2/3" />
      </div>
    </LoadingStatus>
  );
}

export function WorkSurfaceLayoutSkeleton({
  layout,
}: {
  layout: WorkRouteSearch["layout"];
}) {
  switch (layout) {
    case "board":
      return <WorkBoardSkeleton />;
    case "table":
      return <WorkTableSkeleton />;
    case "timeline":
      return <WorkTimelineSkeleton />;
    default:
      return <WorkListSkeleton />;
  }
}

export function WorkItemPageSkeleton() {
  return (
    <ContentWidth
      width="overview"
      className="space-y-7"
      role="status"
      aria-busy="true"
    >
      <span className="sr-only">Loading page</span>
      <div className="space-y-2">
        <Skeleton className="h-3 w-16" />
        <Skeleton className="h-8 w-72" />
        <Skeleton className="h-4 w-96 max-w-full" />
      </div>
      <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="space-y-4">
          <Skeleton className="h-40 w-full rounded-xl" />
          <Skeleton className="h-24 w-full rounded-xl" />
        </div>
        <div className="space-y-4">
          <Skeleton className="h-3 w-16" />
          <div className="divide-border/60 divide-y">
            {Array.from({ length: 5 }, (_, index) => (
              <div
                key={index}
                className="grid grid-cols-[minmax(7rem,0.75fr)_minmax(0,1.25fr)] items-start gap-4 py-3"
              >
                <Skeleton className="h-3 w-16" />
                <Skeleton className="h-4 w-28" />
              </div>
            ))}
          </div>
        </div>
      </div>
    </ContentWidth>
  );
}
