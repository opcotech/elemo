import {
  ArrowDownAZIcon,
  CalendarDaysIcon,
  ChevronDownIcon,
  Columns3Icon,
  FilterIcon,
  GalleryHorizontalEndIcon,
  GroupIcon,
  LayoutListIcon,
  PanelTopIcon,
  SlidersHorizontalIcon,
  Table2Icon,
} from "lucide-react";
import type { ReactNode } from "react";

import { parseSort } from "./utils";
import type { SearchPatch } from "./utils";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import type { SavedView, Scope, WorkLayout } from "@/lib/mock-data";
import {
  resolveWorkScope,
  serializeWorkScope,
  workScopeOptions,
} from "@/lib/work-route-search";
import type { WorkRouteSearch } from "@/lib/work-route-search";

const layouts: readonly {
  value: WorkLayout;
  label: string;
  icon: typeof Columns3Icon;
}[] = [
  { value: "list", label: "List", icon: LayoutListIcon },
  { value: "table", label: "Table", icon: Table2Icon },
  { value: "board", label: "Board", icon: Columns3Icon },
  { value: "timeline", label: "Timeline", icon: CalendarDaysIcon },
];

const groupLabels: Record<WorkRouteSearch["group"], string> = {
  status: "Status",
  priority: "Priority",
  assignee: "Assignee",
  none: "No grouping",
};

const sortLabels: Record<string, string> = {
  "rank:asc": "Manual rank",
  "priority:desc": "Priority",
  "dueDate:asc": "Due date",
  "updatedAt:desc": "Recently updated",
};

const displayLabels: Record<WorkRouteSearch["display"], string> = {
  comfortable: "Comfortable",
  compact: "Compact",
};

function scopeLabel(scope: Scope) {
  if (scope.type === "project") return "This project";
  if (scope.type === "namespace") return "This namespace";
  if (scope.type === "person") return "My related work";
  return "Everywhere";
}

function ViewMenuButton({
  icon: Icon,
  label,
  children,
  align = "start",
  "aria-label": ariaLabel,
}: {
  icon: typeof Columns3Icon;
  label: ReactNode;
  children: ReactNode;
  align?: "start" | "end";
  "aria-label"?: string;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={<Button variant="outline" size="sm" aria-label={ariaLabel} />}
      >
        <Icon aria-hidden />
        {label}
        <ChevronDownIcon />
      </DropdownMenuTrigger>
      <DropdownMenuContent align={align}>{children}</DropdownMenuContent>
    </DropdownMenu>
  );
}

export function ScopePicker({
  baseScope,
  value,
  onValueChange,
}: {
  baseScope: Scope;
  value: WorkRouteSearch["scope"];
  onValueChange: (scope: string) => void;
}) {
  const options = workScopeOptions(baseScope);
  const scope = resolveWorkScope(value, baseScope);

  return (
    <ViewMenuButton
      icon={GalleryHorizontalEndIcon}
      label={<>Scope: {scopeLabel(scope)}</>}
    >
      <DropdownMenuRadioGroup
        value={serializeWorkScope(scope)}
        onValueChange={onValueChange}
      >
        {options.map((option) => (
          <DropdownMenuRadioItem
            key={serializeWorkScope(option)}
            value={serializeWorkScope(option)}
          >
            {scopeLabel(option)}
          </DropdownMenuRadioItem>
        ))}
      </DropdownMenuRadioGroup>
    </ViewMenuButton>
  );
}

export function ViewBar({
  search,
  scope,
  savedViews,
  onSearchChange,
  itemCount,
  showScopePicker = true,
}: {
  search: WorkRouteSearch;
  scope: Scope;
  savedViews: readonly SavedView[];
  onSearchChange: (patch: SearchPatch) => void;
  itemCount: number;
  showScopePicker?: boolean;
}) {
  const activeView = savedViews.find((view) => view.id === search.view);
  const sortValue = parseSort(search.sort);

  const selectView = (viewId: string) => {
    if (viewId === "all") {
      onSearchChange({ view: undefined });
      return;
    }

    const view = savedViews.find((candidate) => candidate.id === viewId);
    if (!view) return;
    const primarySort = view.sort[0];

    onSearchChange({
      view: view.id,
      scope: serializeWorkScope(view.scope),
      layout: view.layout,
      group: view.groupBy,
      sort: primarySort
        ? `${primarySort.field}:${primarySort.direction}`
        : "rank:asc",
    });
  };

  return (
    <div className="bg-background/95 sticky top-0 z-20 space-y-2 border-y px-3 py-2 backdrop-blur sm:px-4">
      <div className="flex flex-wrap items-center gap-2">
        <ViewMenuButton
          icon={PanelTopIcon}
          label={activeView?.name ?? "All work"}
        >
          <DropdownMenuRadioGroup
            value={activeView?.id ?? "all"}
            onValueChange={selectView}
          >
            <DropdownMenuLabel>Illustrative saved views</DropdownMenuLabel>
            <DropdownMenuRadioItem value="all">All work</DropdownMenuRadioItem>
            {savedViews.map((view) => (
              <DropdownMenuRadioItem key={view.id} value={view.id}>
                {view.name}
              </DropdownMenuRadioItem>
            ))}
          </DropdownMenuRadioGroup>
        </ViewMenuButton>

        {showScopePicker ? (
          <ScopePicker
            baseScope={scope}
            value={search.scope}
            onValueChange={(value) => onSearchChange({ scope: value })}
          />
        ) : null}

        <div className="order-last flex w-full items-center rounded-lg border p-0.5 sm:order-0 sm:w-auto">
          {layouts.map((layout) => (
            <Button
              key={layout.value}
              type="button"
              variant={search.layout === layout.value ? "secondary" : "ghost"}
              size="xs"
              className="flex-1 sm:flex-none"
              aria-label={layout.label}
              aria-pressed={search.layout === layout.value}
              onClick={() => onSearchChange({ layout: layout.value })}
            >
              <layout.icon aria-hidden />
              <span className="hidden lg:inline">{layout.label}</span>
            </Button>
          ))}
        </div>

        <div className="ml-auto flex flex-wrap items-center gap-1">
          <DropdownMenu>
            <DropdownMenuTrigger
              render={
                <Button
                  variant={search.filter ? "secondary" : "ghost"}
                  size="sm"
                  aria-label="Filter"
                />
              }
            >
              <FilterIcon aria-hidden />
              <span className="hidden md:inline">Filter</span>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="min-w-72 p-3">
              <div className="text-muted-foreground px-2.5 py-1.5 text-xs font-medium">
                Natural-language filter
              </div>
              <Input
                value={search.filter ?? ""}
                onChange={(event) =>
                  onSearchChange({ filter: event.target.value || undefined })
                }
                placeholder="Key, title, label, status..."
                aria-label="Filter work"
              />
              <p className="text-muted-foreground mt-2 px-1 text-xs">
                Matches work identity, summary, and labels.
              </p>
            </DropdownMenuContent>
          </DropdownMenu>

          <ViewMenuButton
            icon={GroupIcon}
            label={groupLabels[search.group]}
            align="end"
            aria-label="Group by"
          >
            <DropdownMenuRadioGroup
              value={search.group}
              onValueChange={(value) =>
                onSearchChange({
                  group: value as WorkRouteSearch["group"],
                })
              }
            >
              <DropdownMenuRadioItem value="status">
                Status
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="priority">
                Priority
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="assignee">
                Assignee
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="none">
                No grouping
              </DropdownMenuRadioItem>
            </DropdownMenuRadioGroup>
          </ViewMenuButton>

          <ViewMenuButton
            icon={ArrowDownAZIcon}
            label={sortLabels[sortValue] ?? "Sort"}
            align="end"
            aria-label="Sort"
          >
            <DropdownMenuRadioGroup
              value={sortValue}
              onValueChange={(value) => onSearchChange({ sort: value })}
            >
              <DropdownMenuRadioItem value="rank:asc">
                Manual rank
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="priority:desc">
                Priority
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="dueDate:asc">
                Due date
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="updatedAt:desc">
                Recently updated
              </DropdownMenuRadioItem>
            </DropdownMenuRadioGroup>
          </ViewMenuButton>

          <ViewMenuButton
            icon={SlidersHorizontalIcon}
            label={displayLabels[search.display]}
            align="end"
            aria-label="Display density"
          >
            <DropdownMenuRadioGroup
              value={search.display}
              onValueChange={(value) =>
                onSearchChange({
                  display: value as WorkRouteSearch["display"],
                })
              }
            >
              <DropdownMenuRadioItem value="comfortable">
                Comfortable
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="compact">
                Compact
              </DropdownMenuRadioItem>
            </DropdownMenuRadioGroup>
          </ViewMenuButton>
        </div>
        <span className="text-muted-foreground text-xs tabular-nums">
          {itemCount} {itemCount === 1 ? "item" : "items"}
        </span>
      </div>

      {(search.filter || activeView) && (
        <div className="flex items-center gap-2 overflow-x-auto text-xs">
          <span className="text-muted-foreground shrink-0">Active:</span>
          {activeView && (
            <button
              type="button"
              onClick={() => onSearchChange({ view: undefined })}
              className="bg-primary-subtle text-primary-on-subtle rounded-full px-2 py-1"
            >
              View: {activeView.name} ×
            </button>
          )}
          {search.filter && (
            <button
              type="button"
              onClick={() => onSearchChange({ filter: undefined })}
              className="bg-primary-subtle text-primary-on-subtle rounded-full px-2 py-1"
            >
              “{search.filter}” ×
            </button>
          )}
        </div>
      )}
    </div>
  );
}
