import { useInfiniteQuery } from "@tanstack/react-query";
import { PlusIcon, SearchIcon } from "lucide-react";
import {
  Suspense,
  lazy,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";

import type { SearchPatch } from "./utils";
import { selectedWorkId } from "./utils";
import type { WorkFieldOverride } from "./work-field-overrides";

import { ResponsiveInspectorShell } from "@/components/layout/responsive-inspector-shell";
import { openQuickCreate } from "@/components/quick-create/open";
import { MockDataAlert } from "@/components/shared/app-feedback";
import { ContextLine } from "@/components/shared/context-line";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { useBoardIssueMove } from "@/components/work/use-board-issue-move";
import type { BoardItemMove } from "@/components/work/use-board-issue-move";
import { useTimelineIssueDates } from "@/components/work/use-timeline-issue-dates";
import type { TimelineDateChange } from "@/components/work/use-timeline-issue-dates";
import { ViewBar } from "@/components/work/view-bar";
import { CompactWorkList } from "@/components/work/work-list";
import {
  WorkInspectorSkeleton,
  WorkSurfaceLayoutSkeleton,
} from "@/components/work/work-surface-skeletons";
import {
  MAX_CURSOR_PAGES,
  cursorPageQueryWith,
  flattenCursorPages,
  nextCursorPageToken,
} from "@/lib/api/cursor-pages";
import {
  v1NamespacesIssuesGetOptions,
  v1ProjectsIssuesGetOptions,
  v1UsersIssuesGetOptions,
} from "@/lib/api/query-options";
import {
  v1NamespacesIssuesGet,
  v1ProjectsIssuesGet,
  v1UsersIssuesGet,
} from "@/lib/api/sdk";
import type { PartialIssue } from "@/lib/api/types";
import { selectSavedViews, selectWorkItems } from "@/lib/mock-data";
import { issuesToWorkItems } from "@/lib/work/issue-adapter";
import {
  buildIssueListApiQuery,
  issueListClientOnlyFilters,
} from "@/lib/work/issue-list-query";
import type {
  Scope,
  WorkFilters,
  WorkItem,
  WorkPriority,
  WorkSortField,
  WorkStatus,
} from "@/lib/work/model";
import { isWorkItemInScope, queryWorkItems } from "@/lib/work/query";
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
  return isWorkItemInScope(item, scope);
}

function useMockScopedWorkItems({
  effectiveScope,
  search,
  activeViewFilters,
}: {
  effectiveScope: Scope;
  search: WorkRouteSearch;
  activeViewFilters?: WorkFilters;
}) {
  return useMemo(
    () =>
      selectWorkItems({
        scope: effectiveScope,
        filters: {
          ...activeViewFilters,
          ...(search.filter ? { text: search.filter } : {}),
        },
        sort: [parseWorkSort(search.sort)],
      }),
    [activeViewFilters, effectiveScope, search.filter, search.sort]
  );
}

function useListedWorkItems({
  listOptions,
  fetchPage,
  toWorkItems,
  search,
  activeViewFilters,
}: {
  listOptions: {
    queryKey: readonly unknown[];
    staleTime?: unknown;
    gcTime?: unknown;
  };
  fetchPage: (
    pageToken: string | undefined,
    signal: AbortSignal
  ) => Promise<{ items?: readonly PartialIssue[] | null } | null | undefined>;
  toWorkItems: (issues: readonly PartialIssue[]) => WorkItem[];
  search: WorkRouteSearch;
  activeViewFilters?: WorkFilters;
}) {
  const { data, error, isPending, isFetchingNextPage, fetchNextPage } =
    useInfiniteQuery({
      staleTime: listOptions.staleTime as number | undefined,
      gcTime: listOptions.gcTime as number | undefined,
      queryKey: [...listOptions.queryKey, "progressive"],
      initialPageParam: undefined as string | undefined,
      queryFn: ({ pageParam, signal }) => fetchPage(pageParam, signal),
      getNextPageParam: (lastPage) => nextCursorPageToken(lastPage),
    });

  const pages = data?.pages ?? [];
  const lastPage = pages[pages.length - 1];
  const hasMorePages =
    Boolean(nextCursorPageToken(lastPage)) && pages.length < MAX_CURSOR_PAGES;
  const collectionComplete =
    !hasMorePages || (Boolean(error) && pages.length > 0);

  useEffect(() => {
    if (isPending || isFetchingNextPage || !hasMorePages) {
      return;
    }
    if (error && pages.length > 0) {
      return;
    }
    void fetchNextPage();
  }, [
    error,
    fetchNextPage,
    hasMorePages,
    isFetchingNextPage,
    isPending,
    pages.length,
  ]);

  const items = useMemo(() => {
    const mergedItems = flattenCursorPages(pages);
    return queryWorkItems(toWorkItems(mergedItems), {
      filters: issueListClientOnlyFilters(activeViewFilters),
      sort: [parseWorkSort(search.sort)],
    });
  }, [activeViewFilters, pages, search, toWorkItems]);

  const hasFirstPage = pages.length > 0;
  const surfacedError = hasFirstPage ? undefined : error;

  return {
    items,
    error: surfacedError,
    isPending: !hasFirstPage && isPending,
    isLoadingMore: hasMorePages || isFetchingNextPage,
    isCollectionComplete: collectionComplete,
  };
}

function WorkSurfaceBody({
  title,
  description,
  context,
  scope,
  effectiveScope,
  scopedItems,
  usesApiIssues,
  issuesError,
  issuesPending,
  issuesLoadingMore,
  issuesCollectionComplete,
  search,
  onSearchChange,
  savedViews,
}: {
  title: string;
  description?: string;
  context?: { namespace?: string; project?: string };
  scope: Scope;
  effectiveScope: Scope;
  scopedItems: readonly WorkItem[];
  usesApiIssues: boolean;
  issuesError?: unknown;
  issuesPending?: boolean;
  issuesLoadingMore?: boolean;
  issuesCollectionComplete?: boolean;
  search: WorkRouteSearch;
  onSearchChange: (patch: SearchPatch) => void;
  savedViews: ReturnType<typeof selectSavedViews>;
}) {
  const projectId =
    effectiveScope.type === "project" ? effectiveScope.projectId : undefined;
  const { moveIssue } = useBoardIssueMove(projectId);
  const { updateDates } = useTimelineIssueDates(projectId);
  const [fieldOverrides, setFieldOverrides] = useState<
    ReadonlyMap<string, WorkFieldOverride>
  >(() => new Map());

  useEffect(() => {
    setFieldOverrides((previous) => {
      if (previous.size === 0) {
        return previous;
      }

      let changed = false;
      const next = new Map(previous);
      for (const [id, override] of previous) {
        const item = scopedItems.find((candidate) => candidate.id === id);
        if (!item) {
          next.delete(id);
          changed = true;
          continue;
        }

        const statusMatches =
          override.status === undefined || item.status === override.status;
        const priorityMatches =
          override.priority === undefined ||
          item.priority === override.priority;
        const startMatches =
          override.startDate === undefined ||
          item.startDate === override.startDate;
        const dueMatches =
          override.dueDate === undefined || item.dueDate === override.dueDate;
        if (statusMatches && priorityMatches && startMatches && dueMatches) {
          next.delete(id);
          changed = true;
        }
      }

      return changed ? next : previous;
    });
  }, [scopedItems]);

  const displayItems = useMemo(
    () =>
      scopedItems.map((item) => {
        const override = fieldOverrides.get(item.id);
        return override ? { ...item, ...override } : item;
      }),
    [fieldOverrides, scopedItems]
  );

  const selectedId = selectedWorkId(search.selected);
  const selectedCandidate = selectedId
    ? displayItems.find(
        (item) => item.id === selectedId || item.key === selectedId
      )
    : undefined;
  const selectedItem =
    selectedCandidate && workItemMatchesScope(selectedCandidate, effectiveScope)
      ? selectedCandidate
      : undefined;
  const compact = search.display === "compact";
  const selectItem = (item: WorkItem) =>
    onSearchChange({ selected: `work:${item.key}` });
  const closeInspector = () => onSearchChange({ selected: undefined });

  const handleItemMove = (move: BoardItemMove) => {
    if (usesApiIssues) {
      void moveIssue(move);
      return;
    }

    const patch: WorkFieldOverride =
      move.group === "status"
        ? { status: move.to as WorkStatus }
        : { priority: move.to as WorkPriority };

    setFieldOverrides((previous) => {
      const next = new Map(previous);
      next.set(move.item.id, { ...previous.get(move.item.id), ...patch });
      return next;
    });
  };

  const handleDatesChange = (change: TimelineDateChange) => {
    if (usesApiIssues) {
      void updateDates(change);
      return;
    }

    const patch: WorkFieldOverride = {
      startDate: change.startDate,
      dueDate: change.dueDate,
    };

    setFieldOverrides((previous) => {
      const next = new Map(previous);
      next.set(change.item.id, { ...previous.get(change.item.id), ...patch });
      return next;
    });
  };

  return (
    <div
      className="flex h-[calc(100svh-var(--app-header-height))] min-h-0 min-w-0 flex-col overflow-hidden"
      data-section="work-surface"
    >
      <ResponsiveInspectorShell
        className="min-h-0 min-w-0 flex-1"
        inspector={
          selectedItem ? (
            <Suspense fallback={<WorkInspectorSkeleton />}>
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
            itemCount={displayItems.length}
            showScopePicker={!usesApiIssues}
          />
          <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-hidden p-3 sm:p-4">
            {usesApiIssues ? (
              <MockDataAlert
                title="Live issues with illustrative extras"
                className="shrink-0"
              >
                Issue cards and the timeline come from the live issues API.
                Saved views and relationships remain fixture-backed where those
                APIs are not available yet.
              </MockDataAlert>
            ) : (
              <MockDataAlert
                title="Illustrative work projection"
                className="shrink-0"
              >
                Board, List, Table, Timeline, and saved views use clearly
                labeled fixture data. Scope, filters, layout, density, and
                selection remain URL-owned. Open a project Work surface for live
                issues.
              </MockDataAlert>
            )}
            {usesApiIssues && issuesLoadingMore ? (
              <p className="text-muted-foreground shrink-0 text-xs">
                Loading more issues in the background...
              </p>
            ) : null}
            <div className="min-h-0 min-w-0 flex-1 overflow-hidden">
              {issuesPending ? (
                <WorkSurfaceLayoutSkeleton layout={search.layout} />
              ) : issuesError ? (
                <div className="h-full overflow-auto">
                  <EmptyState
                    icon={<SearchIcon />}
                    title="Couldn't load work"
                    description="The issue list could not be loaded. Try again later."
                  />
                </div>
              ) : displayItems.length === 0 ? (
                <div className="h-full overflow-auto">
                  {usesApiIssues && !issuesCollectionComplete ? (
                    <EmptyState
                      icon={<SearchIcon />}
                      title="Loading more work"
                      description="Still collecting additional pages that may match the current filters."
                    />
                  ) : (
                    <EmptyState
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
                            onClick={() =>
                              onSearchChange({ filter: undefined })
                            }
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
                  )}
                </div>
              ) : search.layout === "board" ? (
                <Suspense
                  fallback={<WorkSurfaceLayoutSkeleton layout="board" />}
                >
                  <WorkBoard
                    items={displayItems}
                    group={search.group}
                    compact={compact}
                    onSelect={selectItem}
                    onItemMove={handleItemMove}
                  />
                </Suspense>
              ) : search.layout === "table" ? (
                <Suspense
                  fallback={<WorkSurfaceLayoutSkeleton layout="table" />}
                >
                  <WorkTable
                    items={displayItems}
                    compact={compact}
                    onSelect={selectItem}
                  />
                </Suspense>
              ) : search.layout === "timeline" ? (
                <Suspense
                  fallback={<WorkSurfaceLayoutSkeleton layout="timeline" />}
                >
                  <WorkTimeline
                    items={displayItems}
                    compact={compact}
                    onSelect={selectItem}
                    onDatesChange={handleDatesChange}
                  />
                </Suspense>
              ) : (
                <div className="h-full min-w-0 overflow-auto">
                  <CompactWorkList
                    items={displayItems}
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

function ProjectWorkSurface({
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
  scope: Extract<Scope, { type: "project" }>;
  search: WorkRouteSearch;
  onSearchChange: (patch: SearchPatch) => void;
}) {
  const effectiveScope = resolveWorkScope(search.scope, scope);
  const projectScope =
    effectiveScope.type === "project" ? effectiveScope : scope;
  const savedViews = useMemo(
    () => selectSavedViews({ scope, includeGlobal: true }),
    [scope]
  );
  const activeView = savedViews.find((view) => view.id === search.view);
  const issueListQuery = useMemo(
    () => buildIssueListApiQuery(search, activeView?.filters),
    [activeView?.filters, search]
  );
  const toWorkItems = useCallback(
    (issues: readonly PartialIssue[]) =>
      issuesToWorkItems(issues, {
        namespaceId: projectScope.namespaceId,
        projectId: projectScope.projectId,
      }),
    [projectScope.namespaceId, projectScope.projectId]
  );
  const {
    items: scopedItems,
    error: issuesError,
    isPending: issuesPending,
    isLoadingMore: issuesLoadingMore,
    isCollectionComplete: issuesCollectionComplete,
  } = useListedWorkItems({
    listOptions: v1ProjectsIssuesGetOptions({
      path: { id: projectScope.projectId },
      query: cursorPageQueryWith(issueListQuery),
    }),
    fetchPage: async (pageToken, signal) => {
      const { data } = await v1ProjectsIssuesGet({
        path: { id: projectScope.projectId },
        query: cursorPageQueryWith(issueListQuery, pageToken),
        signal,
        throwOnError: true,
      });
      return data;
    },
    toWorkItems,
    search,
    activeViewFilters: activeView?.filters,
  });

  return (
    <WorkSurfaceBody
      title={title}
      description={description}
      context={context}
      scope={scope}
      effectiveScope={projectScope}
      scopedItems={scopedItems}
      usesApiIssues
      issuesError={issuesError}
      issuesPending={issuesPending}
      issuesLoadingMore={issuesLoadingMore}
      issuesCollectionComplete={issuesCollectionComplete}
      search={search}
      onSearchChange={onSearchChange}
      savedViews={savedViews}
    />
  );
}

function NamespaceWorkSurface({
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
  scope: Extract<Scope, { type: "namespace" }>;
  search: WorkRouteSearch;
  onSearchChange: (patch: SearchPatch) => void;
}) {
  const effectiveScope = resolveWorkScope(search.scope, scope);
  const namespaceScope =
    effectiveScope.type === "namespace" ? effectiveScope : scope;
  const savedViews = useMemo(
    () => selectSavedViews({ scope, includeGlobal: true }),
    [scope]
  );
  const activeView = savedViews.find((view) => view.id === search.view);
  const issueListQuery = useMemo(
    () => buildIssueListApiQuery(search, activeView?.filters),
    [activeView?.filters, search]
  );
  const toWorkItems = useCallback(
    (issues: readonly PartialIssue[]) =>
      issuesToWorkItems(issues, { namespaceId: namespaceScope.namespaceId }),
    [namespaceScope.namespaceId]
  );
  const {
    items: scopedItems,
    error: issuesError,
    isPending: issuesPending,
    isLoadingMore: issuesLoadingMore,
    isCollectionComplete: issuesCollectionComplete,
  } = useListedWorkItems({
    listOptions: v1NamespacesIssuesGetOptions({
      path: { id: namespaceScope.namespaceId },
      query: cursorPageQueryWith(issueListQuery),
    }),
    fetchPage: async (pageToken, signal) => {
      const { data } = await v1NamespacesIssuesGet({
        path: { id: namespaceScope.namespaceId },
        query: cursorPageQueryWith(issueListQuery, pageToken),
        signal,
        throwOnError: true,
      });
      return data;
    },
    toWorkItems,
    search,
    activeViewFilters: activeView?.filters,
  });

  return (
    <WorkSurfaceBody
      title={title}
      description={description}
      context={context}
      scope={scope}
      effectiveScope={namespaceScope}
      scopedItems={scopedItems}
      usesApiIssues
      issuesError={issuesError}
      issuesPending={issuesPending}
      issuesLoadingMore={issuesLoadingMore}
      issuesCollectionComplete={issuesCollectionComplete}
      search={search}
      onSearchChange={onSearchChange}
      savedViews={savedViews}
    />
  );
}

function PersonWorkSurface({
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
  scope: Extract<Scope, { type: "person" }>;
  search: WorkRouteSearch;
  onSearchChange: (patch: SearchPatch) => void;
}) {
  const effectiveScope = resolveWorkScope(search.scope, scope);
  const personScope = effectiveScope.type === "person" ? effectiveScope : scope;
  const savedViews = useMemo(
    () => selectSavedViews({ scope, includeGlobal: true }),
    [scope]
  );
  const activeView = savedViews.find((view) => view.id === search.view);
  const issueListQuery = useMemo(
    () => buildIssueListApiQuery(search, activeView?.filters),
    [activeView?.filters, search]
  );
  const toWorkItems = useCallback(
    (issues: readonly PartialIssue[]) => issuesToWorkItems(issues),
    []
  );
  const {
    items: scopedItems,
    error: issuesError,
    isPending: issuesPending,
    isLoadingMore: issuesLoadingMore,
    isCollectionComplete: issuesCollectionComplete,
  } = useListedWorkItems({
    listOptions: v1UsersIssuesGetOptions({
      path: { id: personScope.personId },
      query: cursorPageQueryWith(issueListQuery),
    }),
    fetchPage: async (pageToken, signal) => {
      const { data } = await v1UsersIssuesGet({
        path: { id: personScope.personId },
        query: cursorPageQueryWith(issueListQuery, pageToken),
        signal,
        throwOnError: true,
      });
      return data;
    },
    toWorkItems,
    search,
    activeViewFilters: activeView?.filters,
  });

  return (
    <WorkSurfaceBody
      title={title}
      description={description}
      context={context}
      scope={scope}
      effectiveScope={personScope}
      scopedItems={scopedItems}
      usesApiIssues
      issuesError={issuesError}
      issuesPending={issuesPending}
      issuesLoadingMore={issuesLoadingMore}
      issuesCollectionComplete={issuesCollectionComplete}
      search={search}
      onSearchChange={onSearchChange}
      savedViews={savedViews}
    />
  );
}

function FixtureWorkSurface({
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
  const scopedItems = useMockScopedWorkItems({
    effectiveScope,
    search,
    activeViewFilters: activeView?.filters,
  });

  return (
    <WorkSurfaceBody
      title={title}
      description={description}
      context={context}
      scope={scope}
      effectiveScope={effectiveScope}
      scopedItems={scopedItems}
      usesApiIssues={false}
      search={search}
      onSearchChange={onSearchChange}
      savedViews={savedViews}
    />
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
  if (scope.type === "project") {
    return (
      <ProjectWorkSurface
        title={title}
        description={description}
        context={context}
        scope={scope}
        search={search}
        onSearchChange={onSearchChange}
      />
    );
  }

  if (scope.type === "namespace") {
    return (
      <NamespaceWorkSurface
        title={title}
        description={description}
        context={context}
        scope={scope}
        search={search}
        onSearchChange={onSearchChange}
      />
    );
  }

  if (scope.type === "person") {
    return (
      <PersonWorkSurface
        title={title}
        description={description}
        context={context}
        scope={scope}
        search={search}
        onSearchChange={onSearchChange}
      />
    );
  }

  return (
    <FixtureWorkSurface
      title={title}
      description={description}
      context={context}
      scope={scope}
      search={search}
      onSearchChange={onSearchChange}
    />
  );
}
