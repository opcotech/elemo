import type { ElemoPluginAPI, PluginGraphNode } from "@elemo/plugin-sdk";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  Button,
  Card,
  CardContent,
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
  SearchInput,
  Spinner,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Textarea,
} from "@elemo/plugin-ui";
import { MoreHorizontal, Pencil, Plus, Trash2, Wallet } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { AccountingPage, pluginBasePath } from "./layout";
import { accountLabel, asNodes, propString } from "./nodes";
import { errorMessage, notifyError, notifySuccess } from "./notify";

export function AccountsPage(props: Record<string, unknown>) {
  const elemo = props.elemo as ElemoPluginAPI;
  const organizationId = String(props.organizationId ?? elemo.scope.id);
  const organizationSlug = String(props.organizationSlug ?? "");
  const [accounts, setAccounts] = useState<PluginGraphNode[]>([]);
  const [budgetCounts, setBudgetCounts] = useState<Record<string, number>>({});
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [edit, setEdit] = useState<PluginGraphNode | null>(null);
  const [remove, setRemove] = useState<PluginGraphNode | null>(null);

  async function reload() {
    const result = asNodes(
      await elemo.api.invoke("account.list", { organizationId }),
    );
    setAccounts(result);
    const counts = await Promise.all(
      result.map(async (account) => {
        const budgets = asNodes(
          await elemo.api.invoke("budget.list", { accountId: account.id }),
        );
        return [account.id, budgets.length] as const;
      }),
    );
    const next: Record<string, number> = {};
    for (const [id, count] of counts) {
      next[id] = count;
    }
    setBudgetCounts(next);
  }

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    void reload()
      .then(() => {
        if (!cancelled) {
          setError("");
        }
      })
      .catch((cause) => {
        if (!cancelled) {
          setError(errorMessage(cause));
          notifyError("Failed to load accounts", cause);
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

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) {
      return accounts;
    }
    return accounts.filter((account) =>
      [
        propString(account, "code"),
        propString(account, "name"),
        propString(account, "description"),
      ]
        .join(" ")
        .toLowerCase()
        .includes(q),
    );
  }, [accounts, search]);

  const removeCount = remove ? (budgetCounts[remove.id] ?? 0) : 0;
  const blocked = removeCount > 0;
  const removeLabel = remove ? accountLabel(remove) : "This account";

  return (
    <AccountingPage
      elemo={elemo}
      organizationSlug={organizationSlug}
      title="Chart of accounts"
      description="Create billable accounts, then attach hour envelopes that projects and work items can be charged against."
      current="accounts"
      testId="accounting-accounts"
      actions={
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <Plus className="size-3.5" />
          New account
        </Button>
      }
    >
      {loading ? (
        <div className="flex min-h-56 items-center justify-center">
          <Spinner />
        </div>
      ) : error ? (
        <EmptyState title="Couldn't load accounts" description={error} />
      ) : accounts.length === 0 ? (
        <EmptyState
          icon={<Wallet />}
          title="No accounts yet"
          description="Add an account to start grouping hour budgets for this organization."
          action={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="size-3.5" />
              New account
            </Button>
          }
        />
      ) : (
        <Card>
          <CardContent className="space-y-4">
            <SearchInput
              value={search}
              onChange={setSearch}
              placeholder="Search accounts"
              aria-label="Search accounts"
            />
            {filtered.length === 0 ? (
              <EmptyState
                compact
                title="No matching accounts"
                description="Try a different code or name."
              />
            ) : (
              <Table containerClassName="border-0 bg-transparent rounded-none">
                <TableHeader>
                  <TableRow>
                    <TableHead>Code</TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Budgets</TableHead>
                    <TableHead className="w-10">
                      <span className="sr-only">Actions</span>
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filtered.map((account) => {
                    const count = budgetCounts[account.id] ?? 0;
                    return (
                      <TableRow
                        key={account.id}
                        data-testid="accounting-account-row"
                      >
                        <TableCell className="font-mono text-sm">
                          {propString(account, "code")}
                        </TableCell>
                        <TableCell>{propString(account, "name")}</TableCell>
                        <TableCell className="text-muted-foreground max-w-96 whitespace-normal">
                          {propString(account, "description") || "—"}
                        </TableCell>
                        <TableCell>
                          <Button
                            variant="ghost"
                            size="sm"
                            nativeButton={false}
                            render={
                              <a
                                href={`${pluginBasePath(organizationSlug, elemo.pluginId)}/budgets?account=${encodeURIComponent(account.id)}`}
                              />
                            }
                          >
                            {count === 1 ? "1 budget" : `${count} budgets`}
                          </Button>
                        </TableCell>
                        <TableCell className="text-right">
                          <AccountMenu
                            onEdit={() => setEdit(account)}
                            onDelete={() => setRemove(account)}
                          />
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}
      <AccountFormDialog
        open={createOpen}
        pending={busy}
        title="New account"
        description="Codes appear on invoices and budget reports."
        submitLabel="Create"
        onOpenChange={setCreateOpen}
        onSubmit={async (code, name, accountDescription) => {
          setBusy(true);
          try {
            await elemo.api.invoke("account.create", {
              organizationId,
              code,
              name,
              description: accountDescription,
              active: true,
            });
            setCreateOpen(false);
            notifySuccess("Account created", `${code} ${name}`.trim());
            await reload();
          } catch (cause) {
            notifyError("Failed to create account", cause);
          } finally {
            setBusy(false);
          }
        }}
      />
      <AccountFormDialog
        open={edit !== null}
        pending={busy}
        title="Edit account"
        description="Update the code or display name."
        submitLabel="Save"
        initialCode={edit ? propString(edit, "code") : ""}
        initialName={edit ? propString(edit, "name") : ""}
        initialDescription={edit ? propString(edit, "description") : ""}
        onOpenChange={(open) => {
          if (!open) {
            setEdit(null);
          }
        }}
        onSubmit={async (code, name, accountDescription) => {
          if (!edit) {
            return;
          }
          setBusy(true);
          try {
            await elemo.api.invoke("account.update", {
              id: edit.id,
              properties: { code, name, description: accountDescription },
            });
            setEdit(null);
            notifySuccess("Account updated", `${code} ${name}`.trim());
            await reload();
          } catch (cause) {
            notifyError("Failed to update account", cause);
          } finally {
            setBusy(false);
          }
        }}
      />
      <AlertDialog
        open={blocked && remove !== null}
        onOpenChange={(open) => {
          if (!open) {
            setRemove(null);
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Account has budgets</AlertDialogTitle>
            <AlertDialogDescription>
              {removeLabel} still has{" "}
              {removeCount === 1 ? "1 budget" : `${removeCount} budgets`}.
              Delete those envelopes first.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (!remove) {
                  return;
                }
                window.location.assign(
                  `${pluginBasePath(organizationSlug, elemo.pluginId)}/budgets?account=${encodeURIComponent(remove.id)}`,
                );
              }}
            >
              View budgets
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <DeleteConfirmationDialog
        open={!blocked && remove !== null}
        onOpenChange={(open) => {
          if (!open) {
            setRemove(null);
          }
        }}
        title="Delete account"
        description={`${removeLabel} will be removed from the chart of accounts.`}
        isPending={busy}
        onConfirm={() => {
          if (!remove) {
            return;
          }
          setBusy(true);
          void elemo.api
            .invoke("account.delete", { id: remove.id })
            .then(async () => {
              const label = accountLabel(remove);
              setRemove(null);
              notifySuccess("Account deleted", label);
              await reload();
            })
            .catch((cause: unknown) => {
              notifyError("Failed to delete account", cause);
            })
            .finally(() => setBusy(false));
        }}
      />
    </AccountingPage>
  );
}

function AccountMenu({
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
            aria-label="Account actions"
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

function AccountFormDialog({
  open,
  pending,
  title,
  description,
  submitLabel,
  initialCode = "",
  initialName = "",
  initialDescription = "",
  onOpenChange,
  onSubmit,
}: {
  open: boolean;
  pending: boolean;
  title: string;
  description: string;
  submitLabel: string;
  initialCode?: string;
  initialName?: string;
  initialDescription?: string;
  onOpenChange: (open: boolean) => void;
  onSubmit: (code: string, name: string, description: string) => Promise<void>;
}) {
  const [code, setCode] = useState(initialCode);
  const [name, setName] = useState(initialName);
  const [accountDescription, setAccountDescription] =
    useState(initialDescription);

  useEffect(() => {
    if (open) {
      setCode(initialCode);
      setName(initialName);
      setAccountDescription(initialDescription);
    }
  }, [initialCode, initialDescription, initialName, open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="account-code">Code</Label>
            <Input
              id="account-code"
              value={code}
              onChange={(event) => setCode(event.target.value)}
              placeholder="CONS"
              autoComplete="off"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="account-name">Name</Label>
            <Input
              id="account-name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="Consulting"
            />
          </div>
        </div>
        <div className="space-y-2">
          <Label htmlFor="account-description">Description</Label>
          <Textarea
            id="account-description"
            value={accountDescription}
            onChange={(event) => setAccountDescription(event.target.value)}
            placeholder="What work is charged to this account?"
            rows={3}
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            disabled={pending || !code.trim() || !name.trim()}
            onClick={() =>
              void onSubmit(code.trim(), name.trim(), accountDescription.trim())
            }
          >
            {submitLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
