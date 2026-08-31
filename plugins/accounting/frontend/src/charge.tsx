import type { ElemoPluginAPI, PluginGraphNode } from "@elemo/plugin-sdk";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  SearchableEntitySelect,
} from "@elemo/plugin-ui";
import { useEffect, useMemo, useState } from "react";

import { formatHours } from "./format";
import { accountLabel, asNodes, budgetLabel, propNumber } from "./nodes";
import { notifyError, notifySuccess } from "./notify";

type BudgetOption = {
  budget: PluginGraphNode;
  account: PluginGraphNode;
};

export function BudgetPicker(props: Record<string, unknown>) {
  const elemo = props.elemo as ElemoPluginAPI;
  const organizationId = String(props.organizationId ?? "");
  const parentId = String(props.parentId ?? "");
  const parentType = String(props.parentType ?? "Project");
  const title = String(props.title ?? "Budget");
  const testId = String(props.testId ?? "accounting-budget-picker");
  const [items, setItems] = useState<BudgetOption[]>([]);
  const [budgetId, setBudgetId] = useState("");
  const [summary, setSummary] = useState<{
    used: number;
    remaining: number;
    seconds: number;
  } | null>(null);
  const [busy, setBusy] = useState(false);

  const options = useMemo(
    () =>
      items.map(({ budget, account }) => ({
        value: budget.id,
        title: budgetLabel(budget),
        description: accountLabel(account),
        searchText: `${accountLabel(account)} ${formatHours(propNumber(budget, "seconds"))}`,
        details: (
          <span className="tabular-nums">
            {formatHours(propNumber(budget, "seconds"))}
          </span>
        ),
      })),
    [items],
  );

  async function reload() {
    const accounts = asNodes(
      await elemo.api.invoke("account.list", { organizationId }),
    );
    const nested = await Promise.all(
      accounts.map(async (account) => {
        const budgets = asNodes(
          await elemo.api.invoke("budget.list", { accountId: account.id }),
        );
        return budgets.map((budget) => ({ budget, account }));
      }),
    );
    setItems(nested.flat());
    const current = (await elemo.api.invoke("charge.get", {
      parentId,
      parentType,
    })) as { budgetId?: string | null };
    const selected = current.budgetId ?? "";
    setBudgetId(selected);
    if (selected) {
      const result = (await elemo.api.invoke("budget.summary", {
        budgetId: selected,
      })) as { used?: number; remaining?: number; seconds?: number };
      setSummary({
        used: Number(result.used ?? 0),
        remaining: Number(result.remaining ?? 0),
        seconds: Number(result.seconds ?? 0),
      });
    } else {
      setSummary(null);
    }
  }

  useEffect(() => {
    if (!parentId || !organizationId) {
      return;
    }
    void reload().catch((error) =>
      notifyError("Failed to load budgets", error),
    );
  }, [elemo, organizationId, parentId, parentType]);

  async function save(next: string) {
    if (next === budgetId) {
      return;
    }
    setBusy(true);
    setBudgetId(next);
    try {
      if (!next) {
        await elemo.api.invoke("charge.clear", { parentId, parentType });
        notifySuccess("Budget cleared", title);
      } else {
        await elemo.api.invoke("charge.set", {
          parentId,
          parentType,
          budgetId: next,
        });
        notifySuccess("Budget assigned", title);
      }
      await reload();
    } catch (error) {
      notifyError("Failed to assign budget", error);
      await reload().catch(() => undefined);
    } finally {
      setBusy(false);
    }
  }

  const emptyOption =
    parentType === "Issue" ? "Use project budget" : "No budget";
  const placeholder =
    parentType === "Issue" ? "Use project budget" : "Select a budget";
  const hint =
    parentType === "Issue"
      ? "Inherits the project budget unless you pick an override."
      : "Assign an hour envelope to this project.";

  return (
    <Card size="sm" data-testid={testId}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        {summary ? null : <CardDescription>{hint}</CardDescription>}
      </CardHeader>
      <CardContent className="space-y-3">
        <div data-testid="accounting-budget-select">
          <SearchableEntitySelect
            options={options}
            value={budgetId || undefined}
            placeholder={placeholder}
            searchPlaceholder="Search budgets…"
            emptyMessage="No budgets found."
            emptyOption={emptyOption}
            disabled={busy}
            onValueChange={(next) => {
              void save(next);
            }}
            aria-label="Assigned budget"
          />
        </div>
        {summary ? (
          <p
            className="text-muted-foreground text-sm"
            data-testid="accounting-budget-remaining"
          >
            Used {formatHours(summary.used)} of {formatHours(summary.seconds)}.{" "}
            {formatHours(summary.remaining)} remaining.
          </p>
        ) : null}
      </CardContent>
    </Card>
  );
}
