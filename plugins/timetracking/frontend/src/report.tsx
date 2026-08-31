import type {
  ElemoPluginAPI,
  PluginIssue,
  PluginUser,
} from "@elemo/plugin-sdk";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  DatePicker,
  EmptyState,
  PageHeader,
  SearchInput,
  SearchableEntitySelect,
  Spinner,
} from "@elemo/plugin-ui";
import { Timer } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";

import {
  TimeEntryList,
  loadEntries,
  useResolvedUsers,
  type TimeEntryView,
} from "./entries";
import { dateKey, formatElapsed, personName } from "./format";

export function TimeReport(props: Record<string, unknown>) {
  const elemo = props.elemo as ElemoPluginAPI;
  const organizationId = String(props.organizationId ?? elemo.scope.id ?? "");
  const [entries, setEntries] = useState<TimeEntryView[]>([]);
  const [issues, setIssues] = useState<Record<string, PluginIssue>>({});
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [workItemId, setWorkItemId] = useState(workItemIdFromSearch);
  const [userFilter, setUserFilter] = useState("");
  const [from, setFrom] = useState<Date | null>(null);
  const [to, setTo] = useState<Date | null>(null);

  const reload = useCallback(() => {
    void (async () => {
      try {
        const next = await loadEntries(elemo, organizationId, "Organization");
        setEntries(next);
        setError(null);
        const parentIds = [
          ...new Set(next.map((entry) => entry.parentId).filter(Boolean)),
        ];
        const resolved = await Promise.all(
          parentIds.map(async (id) => {
            try {
              return await elemo.api.issues.get(id);
            } catch {
              return {
                id,
                key: id,
                title: "Unknown work item",
              } satisfies PluginIssue;
            }
          }),
        );
        const map: Record<string, PluginIssue> = {};
        for (const issue of resolved) {
          map[issue.id] = issue;
        }
        setIssues(map);
      } catch (cause) {
        setError(
          cause instanceof Error ? cause.message : "Failed to load entries",
        );
      } finally {
        setLoading(false);
      }
    })();
  }, [elemo, organizationId]);

  useEffect(() => {
    reload();
  }, [reload]);

  const users = useResolvedUsers(
    elemo,
    entries.map((entry) => entry.userId),
  );

  const filtered = useMemo(
    () =>
      entries.filter((entry) =>
        matchesFilters(entry, {
          search,
          workItemId,
          userFilter,
          from,
          to,
          issues,
          users,
        }),
      ),
    [entries, search, workItemId, userFilter, from, to, issues, users],
  );

  const groups = useMemo(
    () => groupByWorkItem(filtered, issues),
    [filtered, issues],
  );

  const workItemOptions = useMemo(() => {
    const seen = new Map<string, PluginIssue>();
    for (const entry of entries) {
      if (entry.parentId && issues[entry.parentId]) {
        seen.set(entry.parentId, issues[entry.parentId]);
      }
    }
    return [...seen.values()].map((issue) => ({
      value: issue.id,
      title: `${issue.key} ${issue.title}`,
    }));
  }, [entries, issues]);

  const userOptions = useMemo(() => {
    const ids = [
      ...new Set(entries.map((entry) => entry.userId).filter(Boolean)),
    ];
    return ids.map((id) => {
      const user = users[id];
      return {
        value: id,
        title: user ? personName(user) : id,
      };
    });
  }, [entries, users]);

  const totals = useMemo(
    () => ({
      seconds: filtered.reduce((sum, entry) => sum + entry.seconds, 0),
      entries: filtered.length,
      people: new Set(filtered.map((entry) => entry.userId).filter(Boolean))
        .size,
    }),
    [filtered],
  );

  if (loading) {
    return (
      <div data-testid="timetracking-report" className="space-y-6">
        <ReportHeader />
        <div className="flex min-h-56 items-center justify-center">
          <Spinner />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div data-testid="timetracking-report" className="space-y-6">
        <ReportHeader />
        <EmptyState title="Time report" description={error} />
      </div>
    );
  }

  return (
    <div data-testid="timetracking-report" className="space-y-6">
      <ReportHeader />
      {entries.length === 0 ? (
        <EmptyState
          icon={<Timer />}
          title="No time logged"
          description="Start a timer on a work item to create TimeEntry nodes."
        />
      ) : (
        <>
          <div className="grid gap-3 sm:grid-cols-3">
            <MetricCard title="Logged" value={formatElapsed(totals.seconds)} />
            <MetricCard title="Entries" value={String(totals.entries)} />
            <MetricCard title="People" value={String(totals.people)} />
          </div>
          <Card size="sm">
            <CardContent className="flex flex-col gap-3 lg:flex-row lg:flex-wrap lg:items-center">
              <SearchInput
                value={search}
                onChange={setSearch}
                placeholder="Search description, work item, or person"
                aria-label="Search time entries"
                className="min-w-40 max-w-none flex-1"
              />
              <SearchableEntitySelect
                options={workItemOptions}
                value={workItemId || undefined}
                triggerClassName="w-full lg:w-44 lg:shrink-0"
                placeholder="Work item"
                emptyOption="All work items"
                searchPlaceholder="Search work items…"
                onValueChange={setWorkItemId}
              />
              <div className="grid grid-cols-2 gap-2 lg:flex lg:w-auto lg:shrink-0">
                <div className="min-w-0 lg:w-40">
                  <DatePicker
                    date={from}
                    onDateChange={setFrom}
                    placeholder="From"
                    clearable
                    aria-label="From date"
                  />
                </div>
                <div className="min-w-0 lg:w-40">
                  <DatePicker
                    date={to}
                    onDateChange={setTo}
                    placeholder="To"
                    clearable
                    aria-label="To date"
                  />
                </div>
              </div>
              <SearchableEntitySelect
                options={userOptions}
                value={userFilter || undefined}
                triggerClassName="w-full lg:w-44 lg:shrink-0"
                placeholder="Logged by"
                emptyOption="Anyone"
                searchPlaceholder="Search people…"
                onValueChange={setUserFilter}
              />
            </CardContent>
          </Card>
          {groups.length === 0 ? (
            <EmptyState
              compact
              icon={<Timer />}
              title="No matching entries"
              description="Try a different search, work item, person, or date range."
            />
          ) : (
            groups.map((group) => (
              <section
                key={group.parentId || "unknown"}
                className="space-y-2 rounded-xl border p-4"
                data-testid="timetracking-report-group"
              >
                <div className="flex items-baseline justify-between gap-3">
                  <h2 className="font-medium">
                    {group.issue
                      ? `${group.issue.key} ${group.issue.title}`
                      : "Unknown work item"}
                  </h2>
                  <span className="text-muted-foreground tabular-nums">
                    {formatElapsed(group.total)}
                  </span>
                </div>
                <TimeEntryList
                  elemo={elemo}
                  entries={group.entries}
                  projectId={group.issue?.projectId}
                  onChanged={reload}
                />
              </section>
            ))
          )}
        </>
      )}
    </div>
  );
}

function ReportHeader() {
  return (
    <PageHeader
      eyebrow="Time tracking"
      title="Time report"
      description="Review logged time across the organization. Filter by work item, person, or date, then edit, move, or delete entries."
    />
  );
}

function MetricCard({ title, value }: { title: string; value: string }) {
  return (
    <Card size="sm">
      <CardHeader>
        <CardTitle className="text-muted-foreground text-xs font-semibold tracking-wide uppercase">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="text-2xl font-semibold tabular-nums">
        {value}
      </CardContent>
    </Card>
  );
}

function matchesFilters(
  entry: TimeEntryView,
  opts: {
    search: string;
    workItemId: string;
    userFilter: string;
    from: Date | null;
    to: Date | null;
    issues: Record<string, PluginIssue>;
    users: Record<string, PluginUser>;
  },
): boolean {
  if (opts.workItemId && entry.parentId !== opts.workItemId) {
    return false;
  }
  if (opts.userFilter && entry.userId !== opts.userFilter) {
    return false;
  }
  const key = dateKey(entry.createdAt);
  if (opts.from && key && key < dateKey(opts.from.toISOString())) {
    return false;
  }
  if (opts.to && key && key > dateKey(opts.to.toISOString())) {
    return false;
  }
  const q = opts.search.trim().toLowerCase();
  if (!q) {
    return true;
  }
  const issue = opts.issues[entry.parentId];
  const user = opts.users[entry.userId];
  const haystack = [
    entry.note,
    issue?.key,
    issue?.title,
    user ? personName(user) : entry.userId,
  ]
    .filter(Boolean)
    .join(" ")
    .toLowerCase();
  return haystack.includes(q);
}

function groupByWorkItem(
  entries: TimeEntryView[],
  issues: Record<string, PluginIssue>,
): Array<{
  parentId: string;
  issue?: PluginIssue;
  entries: TimeEntryView[];
  total: number;
}> {
  const order: string[] = [];
  const map = new Map<string, TimeEntryView[]>();
  for (const entry of entries) {
    const id = entry.parentId || "";
    if (!map.has(id)) {
      order.push(id);
      map.set(id, []);
    }
    map.get(id)?.push(entry);
  }
  return order.map((parentId) => {
    const group = map.get(parentId) ?? [];
    return {
      parentId,
      issue: issues[parentId],
      entries: group,
      total: group.reduce((sum, entry) => sum + entry.seconds, 0),
    };
  });
}

function workItemIdFromSearch(): string {
  if (typeof window === "undefined") {
    return "";
  }
  return new URLSearchParams(window.location.search).get("workItem") ?? "";
}
