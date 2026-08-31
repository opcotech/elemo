import type { ElemoPluginAPI, PluginGraphNode } from "@elemo/plugin-sdk";
import {
  Badge,
  Button,
  Card,
  CardContent,
  DatePicker,
  DeleteConfirmationDialog,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  EmptyState,
  Input,
  Label,
  Progress,
  ProgressLabel,
  ProgressValue,
  SearchableEntitySelect,
  Spinner,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@elemo/plugin-ui";
import {
  MoreHorizontal,
  Pencil,
  Plus,
  Trash2,
  TriangleAlert,
  Wallet,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import {
  currentYearBounds,
  formatHours,
  formatISODate,
  formatPeriod,
  hoursMinutesFromSeconds,
  parseISODate,
  secondsFromHoursMinutes,
} from "./format";
import { AccountingPage } from "./layout";
import {
  accountLabel,
  asNodes,
  budgetLabel,
  budgetThreshold,
  progressIndicatorClass,
  propNumber,
  propString,
  thresholdReached,
  utilizationPercent,
} from "./nodes";
import { notifyError, notifySuccess } from "./notify";

type BudgetSummary = { used: number; seconds: number };

export function BudgetsPage(props: Record<string, unknown>) {
  const elemo = props.elemo as ElemoPluginAPI;
  const organizationId = String(props.organizationId ?? elemo.scope.id);
  const organizationSlug = String(props.organizationSlug ?? "");
  const search =
    typeof window !== "undefined"
      ? new URLSearchParams(window.location.search)
      : null;
  const initialAccount = search?.get("account") ?? "";
  const [accounts, setAccounts] = useState<PluginGraphNode[]>([]);
  const [accountId, setAccountId] = useState(initialAccount);
  const [budgets, setBudgets] = useState<PluginGraphNode[]>([]);
  const [summaries, setSummaries] = useState<Record<string, BudgetSummary>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [edit, setEdit] = useState<PluginGraphNode | null>(null);
  const [remove, setRemove] = useState<PluginGraphNode | null>(null);

  const selected = useMemo(
    () => accounts.find((account) => account.id === accountId),
    [accountId, accounts],
  );
  const accountOptions = useMemo(
    () =>
      accounts.map((account) => ({
        value: account.id,
        title: accountLabel(account),
        description: propString(account, "code"),
      })),
    [accounts],
  );

  async function reloadAccounts() {
    const items = asNodes(
      await elemo.api.invoke("account.list", { organizationId }),
    );
    setAccounts(items);
    setAccountId((current) => {
      if (current && items.some((item) => item.id === current)) {
        return current;
      }
      return items[0]?.id ?? "";
    });
    return items;
  }

  async function reloadBudgets(id: string) {
    if (!id) {
      setBudgets([]);
      setSummaries({});
      return;
    }
    const items = asNodes(
      await elemo.api.invoke("budget.list", { accountId: id }),
    );
    setBudgets(items);
    const results = await Promise.all(
      items.map(async (budget) => {
        const summary = (await elemo.api.invoke("budget.summary", {
          budgetId: budget.id,
        })) as { used?: number; seconds?: number };
        return [
          budget.id,
          {
            used: Number(summary.used ?? 0),
            seconds: Number(summary.seconds ?? propNumber(budget, "seconds")),
          },
        ] as const;
      }),
    );
    const next: Record<string, BudgetSummary> = {};
    for (const [id, summary] of results) {
      next[id] = summary;
    }
    setSummaries(next);
  }

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    void reloadAccounts()
      .then((items) => {
        if (!cancelled) {
          setError("");
          if (items.length === 0) {
            setLoading(false);
          }
        }
      })
      .catch((cause) => {
        if (!cancelled) {
          setError(
            cause instanceof Error ? cause.message : "Failed to load budgets",
          );
          notifyError("Failed to load budgets", cause);
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [elemo, organizationId]);

  useEffect(() => {
    if (!accountId) {
      setBudgets([]);
      setSummaries({});
      return;
    }
    let cancelled = false;
    setLoading(true);
    void reloadBudgets(accountId)
      .catch((cause) => {
        if (!cancelled) {
          notifyError("Failed to load budgets", cause);
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
  }, [accountId, elemo]);

  const canCreate = Boolean(accountId);

  return (
    <AccountingPage
      elemo={elemo}
      organizationSlug={organizationSlug}
      title="Hour budgets"
      description="Give each account an hour envelope for a period. Projects and work items charge logged time against the assigned envelope."
      current="budgets"
      testId="accounting-budgets"
      actions={
        <Button
          size="sm"
          disabled={!canCreate}
          onClick={() => setCreateOpen(true)}
        >
          <Plus className="size-3.5" />
          New budget
        </Button>
      }
    >
      {loading ? (
        <div className="flex min-h-56 items-center justify-center">
          <Spinner />
        </div>
      ) : error ? (
        <EmptyState title="Couldn't load budgets" description={error} />
      ) : accounts.length === 0 ? (
        <EmptyState
          icon={<Wallet />}
          title="Create an account first"
          description="Hour envelopes belong to an account. Add one on the accounts page, then come back here."
        />
      ) : (
        <div className="space-y-4">
          <Card size="sm">
            <CardContent className="space-y-2">
              <Label htmlFor="budget-account">Account</Label>
              <SearchableEntitySelect
                id="budget-account"
                options={accountOptions}
                value={accountId || undefined}
                placeholder="Select an account"
                searchPlaceholder="Search accounts…"
                emptyMessage="No accounts found."
                onValueChange={setAccountId}
                aria-label="Account"
              />
              {selected ? (
                <p className="text-muted-foreground text-sm">
                  Envelopes on {accountLabel(selected)}.
                </p>
              ) : null}
            </CardContent>
          </Card>
          {budgets.length === 0 ? (
            <EmptyState
              icon={<Wallet />}
              title="No budgets on this account"
              description="Create an hour envelope so projects can bill time here."
              action={
                <Button onClick={() => setCreateOpen(true)}>
                  <Plus className="size-3.5" />
                  New budget
                </Button>
              }
            />
          ) : (
            <Card>
              <CardContent>
                <Table containerClassName="border-0 bg-transparent rounded-none">
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Period</TableHead>
                      <TableHead className="text-right">Threshold</TableHead>
                      <TableHead>Used</TableHead>
                      <TableHead className="w-10">
                        <span className="sr-only">Actions</span>
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {budgets.map((budget) => {
                      const summary = summaries[budget.id];
                      const seconds =
                        summary?.seconds ?? propNumber(budget, "seconds");
                      const used = summary?.used ?? 0;
                      const remaining = Math.max(0, seconds - used);
                      const percent = utilizationPercent(used, seconds);
                      return (
                        <TableRow
                          key={budget.id}
                          data-testid="accounting-budget-row"
                        >
                          <TableCell>
                            <div className="font-medium">
                              {budgetLabel(budget)}
                            </div>
                            <div className="text-muted-foreground text-xs">
                              {formatHours(seconds)} envelope
                            </div>
                          </TableCell>
                          <TableCell className="text-muted-foreground text-sm">
                            {formatPeriod(
                              propString(budget, "period_start"),
                              propString(budget, "period_end"),
                            )}
                          </TableCell>
                          <TableCell className="text-right tabular-nums">
                            {budgetThreshold(budget)}%
                          </TableCell>
                          <TableCell className="min-w-44">
                            <div className="flex items-start gap-2">
                              <Progress
                                value={Math.min(100, percent)}
                                className="w-full min-w-40"
                                indicatorClassName={progressIndicatorClass(
                                  thresholdReached(
                                    percent,
                                    budgetThreshold(budget),
                                  ),
                                )}
                              >
                                <ProgressLabel>
                                  {formatHours(used)} used
                                </ProgressLabel>
                                <ProgressValue>
                                  {() => `${formatHours(remaining)} left`}
                                </ProgressValue>
                              </Progress>
                              {thresholdReached(
                                percent,
                                budgetThreshold(budget),
                              ) ? (
                                <ThresholdWarning />
                              ) : null}
                            </div>
                          </TableCell>
                          <TableCell className="text-right">
                            <BudgetMenu
                              onEdit={() => setEdit(budget)}
                              onDelete={() => setRemove(budget)}
                            />
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}
        </div>
      )}
      <BudgetFormDialog
        open={createOpen}
        pending={busy}
        title="New budget"
        description={
          selected
            ? `Hour envelope on ${accountLabel(selected)}.`
            : "Choose an account, then set the envelope and period."
        }
        submitLabel="Create"
        onOpenChange={setCreateOpen}
        onSubmit={async (values) => {
          setBusy(true);
          try {
            await elemo.api.invoke("budget.create", {
              accountId,
              name: values.name,
              seconds: values.seconds,
              threshold: values.threshold,
              period_start: values.start,
              period_end: values.end,
            });
            setCreateOpen(false);
            notifySuccess("Budget created", values.name);
            await reloadBudgets(accountId);
          } catch (cause) {
            notifyError("Failed to create budget", cause);
          } finally {
            setBusy(false);
          }
        }}
      />
      <BudgetFormDialog
        open={edit !== null}
        pending={busy}
        title="Edit budget"
        description="Update the envelope, name, or period."
        submitLabel="Save"
        initial={
          edit
            ? {
                name: budgetLabel(edit),
                seconds: propNumber(edit, "seconds"),
                threshold: budgetThreshold(edit),
                start: propString(edit, "period_start"),
                end: propString(edit, "period_end"),
              }
            : undefined
        }
        onOpenChange={(open) => {
          if (!open) {
            setEdit(null);
          }
        }}
        onSubmit={async (values) => {
          if (!edit) {
            return;
          }
          setBusy(true);
          try {
            await elemo.api.invoke("budget.update", {
              id: edit.id,
              properties: {
                name: values.name,
                seconds: values.seconds,
                threshold: values.threshold,
                period_start: values.start,
                period_end: values.end,
              },
            });
            setEdit(null);
            notifySuccess("Budget updated", values.name);
            await reloadBudgets(accountId);
          } catch (cause) {
            notifyError("Failed to update budget", cause);
          } finally {
            setBusy(false);
          }
        }}
      />
      <DeleteConfirmationDialog
        open={remove !== null}
        onOpenChange={(open) => {
          if (!open) {
            setRemove(null);
          }
        }}
        title="Delete budget"
        description={`${remove ? budgetLabel(remove) : "This envelope"} will be removed.`}
        consequences={[
          "Projects and work items assigned to this envelope will lose that assignment.",
          "Logged time previously counted here will no longer appear against this budget.",
        ]}
        isPending={busy}
        onConfirm={() => {
          if (!remove) {
            return;
          }
          setBusy(true);
          const label = budgetLabel(remove);
          void elemo.api
            .invoke("budget.delete", { id: remove.id })
            .then(async () => {
              setRemove(null);
              notifySuccess("Budget deleted", label);
              await reloadBudgets(accountId);
            })
            .catch((cause: unknown) => {
              notifyError("Failed to delete budget", cause);
            })
            .finally(() => setBusy(false));
        }}
      />
    </AccountingPage>
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

function BudgetMenu({
  onEdit,
  onDelete,
}: {
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            type="button"
            size="icon-sm"
            variant="ghost"
            aria-label="Budget actions"
          />
        }
      >
        <MoreHorizontal className="size-4" />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={onEdit}>
          <Pencil className="size-3.5" />
          Edit
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem variant="destructive" onClick={onDelete}>
          <Trash2 className="size-3.5" />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function BudgetFormDialog({
  open,
  pending,
  title,
  description,
  submitLabel,
  initial,
  onOpenChange,
  onSubmit,
}: {
  open: boolean;
  pending: boolean;
  title: string;
  description: string;
  submitLabel: string;
  initial?: {
    name: string;
    seconds: number;
    threshold: number;
    start: string;
    end: string;
  };
  onOpenChange: (open: boolean) => void;
  onSubmit: (values: {
    name: string;
    seconds: number;
    threshold: number;
    start: string;
    end: string;
  }) => Promise<void>;
}) {
  const defaults = currentYearBounds();
  const split = hoursMinutesFromSeconds(initial?.seconds ?? 40 * 3600);
  const [name, setName] = useState(initial?.name ?? "");
  const [hours, setHours] = useState(split.hours);
  const [minutes, setMinutes] = useState(split.minutes);
  const [threshold, setThreshold] = useState(initial?.threshold ?? 80);
  const [start, setStart] = useState<Date | null>(
    parseISODate(initial?.start) ?? defaults.start,
  );
  const [end, setEnd] = useState<Date | null>(
    parseISODate(initial?.end) ?? defaults.end,
  );

  useEffect(() => {
    if (!open) {
      return;
    }
    const next = hoursMinutesFromSeconds(initial?.seconds ?? 40 * 3600);
    const bounds = currentYearBounds();
    setName(initial?.name ?? "");
    setHours(next.hours);
    setMinutes(next.minutes);
    setThreshold(initial?.threshold ?? 80);
    setStart(parseISODate(initial?.start) ?? bounds.start);
    setEnd(parseISODate(initial?.end) ?? bounds.end);
  }, [initial, open]);

  const startKey = formatISODate(start);
  const endKey = formatISODate(end);
  const seconds = secondsFromHoursMinutes(hours, minutes);
  const invalidPeriod = Boolean(startKey && endKey && startKey > endKey);
  const canSubmit =
    Boolean(
      name.trim() &&
      startKey &&
      endKey &&
      seconds > 0 &&
      threshold >= 1 &&
      threshold <= 100,
    ) && !invalidPeriod;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <div className="space-y-2">
            <Label htmlFor="budget-name">Name</Label>
            <Input
              id="budget-name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="Q1"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label htmlFor="budget-hours">Hours</Label>
              <Input
                id="budget-hours"
                type="number"
                min={0}
                value={hours}
                onChange={(event) =>
                  setHours(Math.max(0, Number(event.target.value) || 0))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="budget-minutes">Minutes</Label>
              <Input
                id="budget-minutes"
                type="number"
                min={0}
                max={59}
                value={minutes}
                onChange={(event) => {
                  const next = Number(event.target.value) || 0;
                  setMinutes(Math.min(59, Math.max(0, next)));
                }}
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="budget-threshold">Warning threshold (%)</Label>
            <Input
              id="budget-threshold"
              type="number"
              min={1}
              max={100}
              value={threshold}
              onChange={(event) =>
                setThreshold(
                  Math.min(100, Math.max(1, Number(event.target.value) || 1)),
                )
              }
            />
            <p className="text-muted-foreground text-xs">
              Warn when used hours reach this percentage of the envelope.
            </p>
          </div>
          <div className="grid gap-3 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="budget-start">Start</Label>
              <DatePicker
                id="budget-start"
                date={start}
                onDateChange={setStart}
                placeholder="Start date"
                aria-label="Start date"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="budget-end">End</Label>
              <DatePicker
                id="budget-end"
                date={end}
                onDateChange={setEnd}
                placeholder="End date"
                aria-label="End date"
              />
            </div>
          </div>
          {invalidPeriod ? (
            <p className="text-destructive text-sm" role="alert">
              The start date must be on or before the end date.
            </p>
          ) : null}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            disabled={pending || !canSubmit}
            onClick={() =>
              void onSubmit({
                name: name.trim(),
                seconds,
                threshold,
                start: startKey,
                end: endKey,
              })
            }
          >
            {submitLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
