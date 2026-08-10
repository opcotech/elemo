import { PlusIcon, SearchIcon } from "lucide-react";
import { Suspense, lazy, useMemo } from "react";

import type { SearchPatch } from "./utils";
import { selectedWorkId } from "./utils";

import { ResponsiveInspectorShell } from "@/components/layout/responsive-inspector-shell";
import { openQuickCreate } from "@/components/quick-create/open";
import { AppEmptyState, MockDataAlert } from "@/components/shared/app-feedback";
import { ContextLine } from "@/components/shared/context-line";
import { Button } from "@/components/ui/button";
import { ViewBar } from "@/components/work/view-bar";
import { CompactWorkList } from "@/components/work/work-list";
import {
  getWorkItem,
  isInScope,
  selectSavedViews,
  selectWorkItems,
} from "@/lib/mock-data";
import type { Scope, WorkItem, WorkSortField } from "@/lib/mock-data";
import { resolveWorkScope } from "@/lib/work-route-search";
import type { WorkRouteSearch } from "@/lib/work-route-search";

const WorkBoard = lazy(() =>
  import("@/components/work/work-board").then((module) => ({
    default: module.WorkBoard,
  }))
);

const WorkTable = lazy(() =>
  import("@/components/work/work-table").then((module) => ({
    default: module.WorkTable,
  }))
);

const WorkTimeline = lazy(() =>
  import("@/components/work/work-timeline").then((module) => ({
    default: module.WorkTimeline,
  }))
);

const WorkInspector = lazy(() =>
  import("@/components/work/work-inspector").then((module) => ({
    default: module.WorkInspector,
  }))
);

const workSortFields: readonly WorkSortField[] = [
  "rank",
  "title",
  "priority",
  "status",
  "dueDate",
  "createdAt",
  "updatedAt",
];

function parseWorkSort(sort: string) {
  const [requestedField, requestedDirection] = sort.split(":");
  const field = workSortFields.includes(requestedField as WorkSortField)
    ? (requestedField as WorkSortField)
    : "rank";

  return {
    field,
    direction: requestedDirection === "desc" ? "desc" : "asc",
  } as const;
}

function workItemMatchesScope(item: WorkItem, scope: Scope) {
  if (scope.type === "person") return item.assigneeId === scope.personId;
  return isInScope(
    { namespaceId: item.namespaceId, projectId: item.projectId },
    scope
  );
}

export function WorkSurface({
  title,
  description,
  context,
  scope,
  search,
  onSearchChange,
}: {
  title: string;
  description?: string;
  context?: { namespace?: string; project?: string };
  scope: Scope;
  search: WorkRouteSearch;
  onSearchChange: (patch: SearchPatch) => void;
}) {
  const effectiveScope = resolveWorkScope(search.scope, scope);
  const savedViews = useMemo(
    () => selectSavedViews({ scope, includeGlobal: true }),
    [scope]
  );
  const activeView = savedViews.find((view) => view.id === search.view);
  const scopedItems = useMemo(
    () =>
      selectWorkItems({
        scope: effectiveScope,
        filters: {
          ...activeView?.filters,
          ...(search.filter ? { text: search.filter } : {}),
        },
        sort: [parseWorkSort(search.sort)],
      }),
    [activeView, effectiveScope, search.filter, search.sort]
  );
  const selectedId = selectedWorkId(search.selected);
  const selectedCandidate = selectedId ? getWorkItem(selectedId) : undefined;
  const selectedItem =
    selectedCandidate && workItemMatchesScope(selectedCandidate, effectiveScope)
      ? selectedCandidate
      : undefined;
  const compact = search.display === "compact";
  const selectItem = (item: WorkItem) =>
    onSearchChange({ selected: `work:${item.id}` });
  const closeInspector = () => onSearchChange({ selected: undefined });

  return (
    <div className="flex h-[calc(100svh-var(--app-header-height))] min-h-0 min-w-0 flex-col overflow-hidden">
      <ResponsiveInspectorShell
        className="min-h-0 min-w-0 flex-1"
        inspector={
          selectedItem ? (
            <Suspense fallback={<WorkSurfaceFallback label="inspector" />}>
              <WorkInspector item={selectedItem} />
            </Suspense>
          ) : undefined
        }
        inspectorTitle={selectedItem?.key}
        inspectorDescription={selectedItem?.title}
        open={Boolean(selectedItem)}
        onOpenChange={(open) => {
          if (!open) closeInspector();
        }}
      >
        <div className="flex h-full min-h-0 min-w-0 flex-col overflow-hidden">
          <div className="flex shrink-0 items-start gap-4 px-4 py-4 sm:px-6">
            <div className="min-w-0 flex-1">
              <h1 className="text-xl font-semibold tracking-tight">{title}</h1>
              {description && (
                <p className="text-muted-foreground mt-1 text-sm">
                  {description}
                </p>
              )}
              <ContextLine
                namespace={context?.namespace}
                project={context?.project}
              />
            </div>
            <Button size="sm" onClick={() => openQuickCreate("work")}>
              <PlusIcon />
              Create
            </Button>
          </div>
          <ViewBar
            search={search}
            scope={scope}
            savedViews={savedViews}
            onSearchChange={onSearchChange}
            itemCount={scopedItems.length}
          />
          <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-hidden p-3 sm:p-4">
            <MockDataAlert
              title="Illustrative work projection"
              className="shrink-0"
            >
              Board, List, Table, Timeline, and saved views use clearly labeled
              fixture data. Scope, filters, layout, density, and selection
              remain URL-owned.
            </MockDataAlert>
            <div className="min-h-0 min-w-0 flex-1 overflow-hidden">
              {scopedItems.length === 0 ? (
                <div className="h-full overflow-auto">
                  <AppEmptyState
                    icon={<SearchIcon />}
                    title={search.filter ? "No query matches" : "No work yet"}
                    description={
                      search.filter
                        ? "The current filter does not match work in this scope."
                        : "Create work in this context to get started."
                    }
                    action={
                      search.filter ? (
                        <Button
                          variant="outline"
                          onClick={() => onSearchChange({ filter: undefined })}
                        >
                          Clear filter
                        </Button>
                      ) : (
                        <Button onClick={() => openQuickCreate("work")}>
                          Quick create
                        </Button>
                      )
                    }
                  />
                </div>
              ) : search.layout === "board" ? (
                <Suspense fallback={<WorkSurfaceFallback label="board" />}>
                  <WorkBoard
                    items={scopedItems}
                    group={search.group}
                    compact={compact}
                    onSelect={selectItem}
                  />
                </Suspense>
              ) : search.layout === "table" ? (
                <Suspense fallback={<WorkSurfaceFallback label="table" />}>
                  <WorkTable
                    items={scopedItems}
                    compact={compact}
                    onSelect={selectItem}
                  />
                </Suspense>
              ) : search.layout === "timeline" ? (
                <Suspense fallback={<WorkSurfaceFallback label="timeline" />}>
                  <WorkTimeline
                    items={scopedItems}
                    scope={effectiveScope}
                    compact={compact}
                    onSelect={selectItem}
                  />
                </Suspense>
              ) : (
                <div className="h-full min-w-0 overflow-auto">
                  <CompactWorkList
                    items={scopedItems}
                    compact={compact}
                    onSelect={selectItem}
                  />
                </div>
              )}
            </div>
          </div>
        </div>
      </ResponsiveInspectorShell>
    </div>
  );
}

function WorkSurfaceFallback({ label }: { label: string }) {
  return (
    <div
      role="status"
      className="text-muted-foreground flex h-full min-h-32 items-center justify-center text-sm"
    >
      Loading {label}…
    </div>
  );
}
