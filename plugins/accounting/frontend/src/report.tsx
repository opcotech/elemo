import type {
  ElemoPluginAPI,
  PluginGraphNode,
  PluginIssue,
  PluginUser,
} from "@elemo/plugin-sdk";
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  DatePicker,
  EmptyState,
  PersonAvatarStack,
  Progress,
  ProgressLabel,
  ProgressValue,
  SearchInput,
  SearchableEntitySelect,
  Spinner,
  Table,
  TableBody,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
  cn,
} from "@elemo/plugin-ui";
import { ChevronRight, CircleAlert, TriangleAlert, Wallet } from "lucide-react";
import { useEffect, useMemo, useState, type ReactNode } from "react";

import { dateKey, formatDateTime, formatHours, formatPeriod } from "./format";
import { AccountingPage, pluginBasePath } from "./layout";
import {
  accountLabel,
  budgetLabel,
  budgetThreshold,
  progressIndicatorClass,
  propNumber,
  propString,
  thresholdReached,
  utilizationPercent,
} from "./nodes";
import { notifyError } from "./notify";

type BudgetReport = {
  budget: PluginGraphNode;
  entries: PluginGraphNode[];
  seconds: number;
  used: number;
  remaining: number;
};

type AccountReport = {
  account: PluginGraphNode;
  budgets: BudgetReport[];
};

type ReportResponse = {
  accounts?: AccountReport[];
};

type TimelogView = {
  id: string;
  createdAt: string | null;
  seconds: number;
  budget: PluginGraphNode;
  note: string;
  parentId: string;
  userId: string;
};

type AccountView = {
  id: string;
  label: string;
  description: string;
  seconds: number;
  used: number;
  remaining: number;
  budgets: BudgetReport[];
  entries: TimelogView[];
};

export function ReportPage(props: Record<string, unknown>) {
  const elemo = props.elemo as ElemoPluginAPI;
  const organizationId = String(props.organizationId ?? elemo.scope.id);
  const organizationSlug = String(props.organizationSlug ?? "");
  const accountsPath = `${pluginBasePath(organizationSlug, elemo.pluginId)}/accounts`;
  const [accounts, setAccounts] = useState<AccountView[]>([]);
  const [issues, setIssues] = useState<Record<string, PluginIssue>>({});
  const [users, setUsers] = useState<Record<string, PluginUser>>({});
  const [search, setSearch] = useState("");
  const [accountId, setAccountId] = useState("");
  const [userFilter, setUserFilter] = useState("");
  const [from, setFrom] = useState<Date | null>(null);
  const [to, setTo] = useState<Date | null>(null);
  const [expanded, setExpanded] = useState<Set<string>>(() => new Set());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    void (async () => {
      const response = (await elemo.api.invoke("report.get", {
        organizationId,
      })) as ReportResponse;
      const nextAccounts = toAccountViews(response.accounts ?? []);
      const entries = nextAccounts.flatMap((account) => account.entries);
      const issueIds = [
        ...new Set(entries.map((entry) => entry.parentId)),
      ].filter(Boolean);
      const userIds = [...new Set(entries.map((entry) => entry.userId))].filter(
        Boolean,
      );
      const [resolvedIssues, resolvedUsers] = await Promise.all([
        Promise.all(
          issueIds.map(async (id) => {
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
        ),
        Promise.all(
          userIds.map(async (id) => {
            try {
              return await elemo.api.users.get(id);
            } catch {
              return {
                id,
                first_name: id,
                last_name: "",
                picture: null,
              } satisfies PluginUser;
            }
          }),
        ),
      ]);
      if (cancelled) {
        return;
      }
      setAccounts(nextAccounts);
      setIssues(
        Object.fromEntries(resolvedIssues.map((issue) => [issue.id, issue])),
      );
      setUsers(
        Object.fromEntries(resolvedUsers.map((user) => [user.id, user])),
      );
      setError("");
    })()
      .catch((cause) => {
        if (!cancelled) {
          const message =
            cause instanceof Error ? cause.message : "Failed to load report";
          setError(message);
          notifyError("Failed to load report", cause);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [elemo, organizationId]);

  const reportAccounts = useMemo(
    () => accounts.filter((account) => account.budgets.length > 0),
    [accounts],
  );

  const accountOptions = useMemo(
    () =>
      reportAccounts.map((account) => ({
        value: account.id,
        title: account.label,
      })),
    [reportAccounts],
  );

  const userOptions = useMemo(() => {
    const ids = [
      ...new Set(
        reportAccounts.flatMap((account) =>
          account.entries.map((entry) => entry.userId),
        ),
      ),
    ].filter(Boolean);
    return ids.map((id) => {
      const user = users[id];
      return {
        value: id,
        title: user ? personName(user) : id,
        avatarSrc: user?.picture,
        avatarFallback: user ? personInitials(user) : id.slice(0, 2),
      };
    });
  }, [reportAccounts, users]);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    const selectedAccountId = reportAccounts.some(
      (account) => account.id === accountId,
    )
      ? accountId
      : "";
    const entryScoped = Boolean(from || to || userFilter);
    return reportAccounts
      .map((account) => {
        if (selectedAccountId && account.id !== selectedAccountId) {
          return null;
        }
        let entries = account.entries;
        if (userFilter) {
          entries = entries.filter((entry) => entry.userId === userFilter);
        }
        if (from || to) {
          entries = entries.filter((entry) =>
            entryInRange(entry.createdAt, from, to),
          );
        }
        if (entryScoped && entries.length === 0) {
          return null;
        }
        const used = entryScoped
          ? entries.reduce((sum, entry) => sum + entry.seconds, 0)
          : account.used;
        const remaining = account.seconds - used;
        return { ...account, entries, used, remaining };
      })
      .filter((account): account is AccountView => {
        if (!account) {
          return false;
        }
        if (!q) {
          return true;
        }
        const entryText = account.entries
          .flatMap((entry) => {
            const issue = issues[entry.parentId];
            const user = users[entry.userId];
            return [
              entry.note,
              issue?.key,
              issue?.title,
              user ? personName(user) : entry.userId,
            ];
          })
          .join(" ");
        const budgetText = account.budgets
          .map(({ budget }) => budgetLabel(budget))
          .join(" ");
        return `${account.label} ${account.description} ${budgetText} ${entryText}`
          .toLowerCase()
          .includes(q);
      });
  }, [accountId, from, issues, reportAccounts, search, to, userFilter, users]);

  const totals = useMemo(
    () =>
      filtered.reduce(
        (acc, account) => ({
          seconds: acc.seconds + account.seconds,
          used: acc.used + account.used,
          remaining: acc.remaining + account.remaining,
        }),
        { seconds: 0, used: 0, remaining: 0 },
      ),
    [filtered],
  );
  const utilizationPercentTotal = utilizationPercent(
    totals.used,
    totals.seconds,
  );

  function selectAccount(value: string) {
    setAccountId(
      reportAccounts.some((account) => account.id === value) ? value : "",
    );
  }

  function toggleAccount(id: string) {
    setExpanded((current) => {
      const next = new Set(current);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }

  return (
    <AccountingPage
      elemo={elemo}
      organizationSlug={organizationSlug}
      title="Account report"
      description="Compare account utilization and inspect the timelogs counted against each budget."
      current="report"
      testId="accounting-report"
    >
      {loading ? (
        <div className="flex min-h-56 items-center justify-center">
          <Spinner />
        </div>
      ) : error ? (
        <EmptyState
          icon={<CircleAlert />}
          title="Couldn't load report"
          description={error}
        />
      ) : reportAccounts.length === 0 ? (
        <EmptyState
          icon={<Wallet />}
          title="No budgets to report"
          description="Create accounts and hour envelopes to see usage here."
          action={
            <Button
              nativeButton={false}
              render={<a href={accountsPath} />}
              size="sm"
            >
              Go to accounts
            </Button>
          }
        />
      ) : (
        <div className="space-y-6">
          <div
            className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4"
            data-testid="accounting-report-summary"
          >
            <MetricCard title="Allocated" value={formatHours(totals.seconds)} />
            <MetricCard title="Used" value={formatHours(totals.used)} />
            <MetricCard
              title="Remaining"
              value={formatHours(Math.max(0, totals.remaining))}
            />
            <MetricCard
              title="Utilization"
              value={`${utilizationPercentTotal}%`}
              badge={
                utilizationPercentTotal > 100 ? (
                  <Badge variant="destructive">Over budget</Badge>
                ) : utilizationPercentTotal >= 90 ? (
                  <Badge variant="warning">Near limit</Badge>
                ) : null
              }
            />
          </div>

          <div
            className="flex flex-nowrap items-center gap-2 overflow-x-auto pb-1"
            data-testid="accounting-report-filters"
          >
            <SearchInput
              value={search}
              onChange={setSearch}
              placeholder="Search accounts, budgets, or timelogs"
              aria-label="Search account report"
              className="min-w-64 max-w-none flex-1"
            />
            <SearchableEntitySelect
              options={accountOptions}
              value={accountId || undefined}
              triggerClassName="w-44 shrink-0"
              placeholder="Account"
              emptyOption="All accounts"
              searchPlaceholder="Search accounts…"
              onValueChange={selectAccount}
              aria-label="Filter by account"
            />
            <div className="w-40 shrink-0">
              <DatePicker
                date={from}
                onDateChange={setFrom}
                placeholder="From"
                clearable
                aria-label="From date"
              />
            </div>
            <div className="w-40 shrink-0">
              <DatePicker
                date={to}
                onDateChange={setTo}
                placeholder="To"
                clearable
                aria-label="To date"
              />
            </div>
            <SearchableEntitySelect
              options={userOptions}
              value={userFilter || undefined}
              triggerClassName="w-44 shrink-0"
              placeholder="Logged by"
              emptyOption="Anyone"
              searchPlaceholder="Search people…"
              onValueChange={setUserFilter}
              aria-label="Filter by person"
            />
          </div>

          {filtered.length === 0 ? (
            <EmptyState
              compact
              title="No matching accounts"
              description="Try a different search, account, person, or date range."
            />
          ) : (
            <div className="space-y-3">
              <h2 className="text-sm font-semibold tracking-wide uppercase">
                Account ledger
              </h2>
              <AccountLedger
                accounts={filtered}
                expanded={expanded}
                issues={issues}
                organizationSlug={organizationSlug}
                users={users}
                onToggle={toggleAccount}
              />
            </div>
          )}
        </div>
      )}
    </AccountingPage>
  );
}

function MetricCard({
  title,
  value,
  badge,
}: {
  title: string;
  value: string;
  badge?: ReactNode;
}) {
  return (
    <Card size="sm">
      <CardHeader>
        <CardTitle className="text-muted-foreground text-xs font-semibold tracking-wide uppercase">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="flex items-center gap-2 text-2xl font-semibold tabular-nums">
        {value}
        {badge}
      </CardContent>
    </Card>
  );
}

function AccountLedger({
  accounts,
  expanded,
  issues,
  organizationSlug,
  users,
  onToggle,
}: {
  accounts: AccountView[];
  expanded: Set<string>;
  issues: Record<string, PluginIssue>;
  organizationSlug: string;
  users: Record<string, PluginUser>;
  onToggle: (id: string) => void;
}) {
  return (
    <div className="space-y-3">
      {accounts.map((account) => {
        const isExpanded = expanded.has(account.id);
        const percent = utilizationPercent(account.used, account.seconds);
        const reached = account.budgets.some(({ budget, used, seconds }) =>
          thresholdReached(
            utilizationPercent(used, seconds),
            budgetThreshold(budget),
          ),
        );
        return (
          <section
            key={account.id}
            data-testid="accounting-report-account-row"
            aria-expanded={isExpanded}
            className="space-y-3 rounded-xl border p-4"
          >
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <button
                type="button"
                className="hover:text-foreground/80 flex min-w-0 flex-1 items-start gap-2 text-left"
                onClick={() => onToggle(account.id)}
                aria-label={`${isExpanded ? "Collapse" : "Expand"} ${account.label} usage`}
              >
                <ChevronRight
                  className={cn(
                    "mt-0.5 size-4 shrink-0 transition-transform",
                    isExpanded && "rotate-90",
                  )}
                />
                <div className="min-w-0">
                  <div className="font-medium">{account.label}</div>
                  <div className="text-muted-foreground text-xs">
                    {account.budgets.length}{" "}
                    {account.budgets.length === 1 ? "budget" : "budgets"}
                  </div>
                  {account.description ? (
                    <div className="text-muted-foreground mt-1 max-w-2xl text-sm">
                      {account.description}
                    </div>
                  ) : null}
                </div>
              </button>
              <div className="flex min-w-48 items-start gap-3 sm:w-80">
                <Progress
                  value={Math.min(100, percent)}
                  className="min-w-0 flex-1"
                  indicatorClassName={progressIndicatorClass(reached)}
                >
                  <ProgressLabel className="font-normal tabular-nums">
                    {formatHours(account.used)} of{" "}
                    {formatHours(account.seconds)}
                  </ProgressLabel>
                  <ProgressValue>
                    {() => `${formatHours(account.remaining)} left`}
                  </ProgressValue>
                </Progress>
                {reached ? <ThresholdWarning /> : null}
              </div>
            </div>
            {isExpanded ? (
              <div data-testid="accounting-report-account-details">
                <AccountDetails
                  account={account}
                  issues={issues}
                  organizationSlug={organizationSlug}
                  users={users}
                />
              </div>
            ) : null}
          </section>
        );
      })}
    </div>
  );
}

function AccountDetails({
  account,
  issues,
  organizationSlug,
  users,
}: {
  account: AccountView;
  issues: Record<string, PluginIssue>;
  organizationSlug: string;
  users: Record<string, PluginUser>;
}) {
  const entries = [...account.entries].sort((left, right) =>
    String(right.createdAt ?? "").localeCompare(String(left.createdAt ?? "")),
  );
  return (
    <div className="space-y-4 border-t pt-4">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Budget</TableHead>
            <TableHead className="text-right">Allocated</TableHead>
            <TableHead className="text-right">Used</TableHead>
            <TableHead className="text-right">Remaining</TableHead>
            <TableHead className="text-right">Threshold</TableHead>
            <TableHead className="min-w-44">Utilization</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {account.budgets.map(({ budget, used, seconds }) => {
            const percent = utilizationPercent(used, seconds);
            const reached = thresholdReached(percent, budgetThreshold(budget));
            return (
              <TableRow key={budget.id}>
                <TableCell>
                  <div className="font-medium">{budgetLabel(budget)}</div>
                  <div className="text-muted-foreground text-xs">
                    {formatPeriod(
                      propString(budget, "period_start"),
                      propString(budget, "period_end"),
                    )}
                  </div>
                </TableCell>
                <TableCell className="text-right tabular-nums">
                  {formatHours(seconds)}
                </TableCell>
                <TableCell className="text-right tabular-nums">
                  {formatHours(used)}
                </TableCell>
                <TableCell className="text-right tabular-nums">
                  {formatHours(Math.max(0, seconds - used))}
                </TableCell>
                <TableCell className="text-right tabular-nums">
                  {budgetThreshold(budget)}%
                </TableCell>
                <TableCell className="min-w-44">
                  <div className="flex items-start gap-2">
                    <Progress
                      value={Math.min(100, percent)}
                      className="w-full min-w-40"
                      indicatorClassName={progressIndicatorClass(reached)}
                    >
                      <ProgressValue>{() => `${percent}%`}</ProgressValue>
                    </Progress>
                    {reached ? <ThresholdWarning /> : null}
                  </div>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>

      {entries.length === 0 ? (
        <EmptyState
          compact
          title="No counted timelogs"
          description="Timelogs assigned to this account's budgets will appear here."
        />
      ) : (
        <Table data-testid="accounting-report-timelogs">
          <TableHeader>
            <TableRow>
              <TableHead>Date</TableHead>
              <TableHead>Budget</TableHead>
              <TableHead>Work item</TableHead>
              <TableHead>Person</TableHead>
              <TableHead>Description</TableHead>
              <TableHead className="text-right">Time</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {entries.map((entry) => {
              const issue = issues[entry.parentId];
              const user = users[entry.userId];
              return (
                <TableRow key={entry.id}>
                  <TableCell className="whitespace-nowrap">
                    {formatDateTime(entry.createdAt)}
                  </TableCell>
                  <TableCell>{budgetLabel(entry.budget)}</TableCell>
                  <TableCell className="max-w-64 whitespace-normal">
                    <WorkItemCell
                      issue={issue}
                      organizationSlug={organizationSlug}
                      fallback={entry.parentId || "Unknown"}
                    />
                  </TableCell>
                  <TableCell>
                    <PersonAvatarStack
                      size="sm"
                      showNames
                      emptyLabel="Unknown"
                      people={
                        user
                          ? [
                              {
                                id: user.id,
                                name: personName(user),
                                picture: user.picture,
                              },
                            ]
                          : []
                      }
                    />
                  </TableCell>
                  <TableCell className="max-w-80 whitespace-normal">
                    {entry.note || "No description"}
                  </TableCell>
                  <TableCell className="text-right font-medium whitespace-nowrap tabular-nums">
                    {formatHours(entry.seconds)}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
          <TableFooter>
            <TableRow>
              <TableCell colSpan={5}>Total counted time</TableCell>
              <TableCell className="text-right tabular-nums">
                {formatHours(account.used)}
              </TableCell>
            </TableRow>
          </TableFooter>
        </Table>
      )}
    </div>
  );
}

function WorkItemCell({
  issue,
  organizationSlug,
  fallback,
}: {
  issue?: PluginIssue;
  organizationSlug: string;
  fallback: string;
}) {
  if (!issue) {
    return fallback;
  }
  const content = (
    <>
      <span className="font-mono text-xs">{issue.key}</span>{" "}
      <span>{issue.title}</span>
    </>
  );
  const href = workItemHref(organizationSlug, issue);
  if (!href) {
    return content;
  }
  return (
    <a
      href={href}
      className="hover:text-primary underline-offset-2 hover:underline"
    >
      {content}
    </a>
  );
}

function ThresholdWarning() {
  return (
    <Badge
      variant="warning"
      data-testid="accounting-threshold-warning"
      className="shrink-0"
    >
      <TriangleAlert />
      Threshold
    </Badge>
  );
}

function workItemHref(
  organizationSlug: string,
  issue: PluginIssue,
): string | null {
  if (!organizationSlug || !issue.namespaceSlug || !issue.key) {
    return null;
  }
  return `/work/${organizationSlug}/${issue.namespaceSlug}/${issue.key}`;
}

function entryInRange(
  createdAt: string | null,
  from: Date | null,
  to: Date | null,
): boolean {
  const key = dateKey(createdAt);
  if (!key) {
    return !from && !to;
  }
  if (from && key < dateKey(from.toISOString())) {
    return false;
  }
  if (to && key > dateKey(to.toISOString())) {
    return false;
  }
  return true;
}

function toAccountViews(accounts: AccountReport[]): AccountView[] {
  return accounts.map(({ account, budgets }) => {
    const entries = budgets.flatMap(({ budget, entries: budgetEntries }) =>
      budgetEntries.map((entry) => ({
        id: entry.id,
        budget,
        seconds: propNumber(entry, "seconds"),
        note: propString(entry, "note"),
        parentId: String(entry.parent_id ?? ""),
        userId: propString(entry, "user_id"),
        createdAt: entry.created_at ?? null,
      })),
    );
    const totals = budgets.reduce(
      (acc, budget) => ({
        seconds: acc.seconds + Number(budget.seconds ?? 0),
        used: acc.used + Number(budget.used ?? 0),
        remaining: acc.remaining + Number(budget.remaining ?? 0),
      }),
      { seconds: 0, used: 0, remaining: 0 },
    );
    return {
      id: account.id,
      label: accountLabel(account),
      description: propString(account, "description"),
      budgets,
      entries,
      ...totals,
    };
  });
}

function personName(user: PluginUser): string {
  return `${user.first_name} ${user.last_name}`.trim() || user.id;
}

function personInitials(user: PluginUser): string {
  return [user.first_name, user.last_name]
    .filter(Boolean)
    .map((part) => part.charAt(0))
    .join("")
    .slice(0, 2)
    .toUpperCase();
}
