import type {
  ElemoPluginAPI,
  PluginGraphNode,
  PluginUser,
} from "@elemo/plugin-sdk";
import {
  Button,
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
  Input,
  Label,
  PersonAvatarStack,
  SearchableEntitySelect,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Textarea,
  cn,
} from "@elemo/plugin-ui";
import { MoreHorizontal, Move, Pencil, Trash2 } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";

import {
  formatDate,
  formatElapsed,
  hoursMinutesFromSeconds,
  personName,
  secondsFromHoursMinutes,
} from "./format";
import { notifyError, notifySuccess } from "./notify";

export interface TimeEntryView {
  id: string;
  seconds: number;
  note: string;
  userId: string;
  parentId: string;
  createdAt: string | null;
}

export function toTimeEntry(node: PluginGraphNode): TimeEntryView {
  const properties = node.properties ?? {};
  return {
    id: node.id,
    seconds: Number(properties.seconds ?? 0),
    note: String(properties.note ?? ""),
    userId: String(properties.user_id ?? ""),
    parentId: String(node.parent_id ?? ""),
    createdAt: node.created_at ?? null,
  };
}

export async function loadEntries(
  elemo: ElemoPluginAPI,
  scopeId: string,
  scopeType: string
): Promise<TimeEntryView[]> {
  const nodes = await elemo.api.graph.nodes.list({
    kind: "TimeEntry",
    scopeId,
    scopeType,
  });
  return nodes.map(toTimeEntry);
}

export function useIssueEntries(
  elemo: ElemoPluginAPI,
  issueId: string
): {
  entries: TimeEntryView[];
  error: string | null;
  reload: () => void;
} {
  const [entries, setEntries] = useState<TimeEntryView[]>([]);
  const [error, setError] = useState<string | null>(null);

  const reload = useCallback(() => {
    if (!issueId) {
      return;
    }
    void loadEntries(elemo, issueId, "Issue")
      .then((next) => {
        setEntries(next);
        setError(null);
      })
      .catch((cause: unknown) => {
        setError(
          cause instanceof Error ? cause.message : "Failed to load time entries"
        );
      });
  }, [elemo, issueId]);

  useEffect(() => {
    reload();
    const timer = window.setInterval(reload, 4000);
    const onVisible = () => {
      if (document.visibilityState === "visible") {
        reload();
      }
    };
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      window.clearInterval(timer);
      document.removeEventListener("visibilitychange", onVisible);
    };
  }, [reload]);

  return { entries, error, reload };
}

export function useResolvedUsers(
  elemo: ElemoPluginAPI,
  userIds: string[]
): Record<string, PluginUser> {
  const [users, setUsers] = useState<Record<string, PluginUser>>({});
  const uniqueIds = useMemo(
    () => [...new Set(userIds.filter(Boolean))],
    [userIds.join(",")]
  );

  useEffect(() => {
    let cancelled = false;
    const missing = uniqueIds.filter((id) => !(id in users));
    if (missing.length === 0) {
      return;
    }
    void Promise.all(
      missing.map(async (id) => {
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
      })
    ).then((resolved) => {
      if (cancelled) {
        return;
      }
      setUsers((current) => {
        const next = { ...current };
        for (const user of resolved) {
          next[user.id] = user;
        }
        return next;
      });
    });
    return () => {
      cancelled = true;
    };
  }, [elemo, uniqueIds.join(",")]);

  return users;
}

export function TimeEntryList({
  elemo,
  entries,
  projectId,
  compact = false,
  onChanged,
}: {
  elemo: ElemoPluginAPI;
  entries: TimeEntryView[];
  projectId?: string;
  compact?: boolean;
  onChanged: () => void;
}) {
  const users = useResolvedUsers(
    elemo,
    entries.map((entry) => entry.userId)
  );
  const [edit, setEdit] = useState<TimeEntryView | null>(null);
  const [move, setMove] = useState<TimeEntryView | null>(null);
  const [remove, setRemove] = useState<TimeEntryView | null>(null);
  const [pending, setPending] = useState(false);
  const cellClass = compact ? "px-2 py-1.5" : undefined;

  return (
    <>
      <Table>
        <TableHeader className={compact ? "sr-only" : undefined}>
          <TableRow>
            <TableHead>Logged by</TableHead>
            <TableHead className={compact ? "w-full" : undefined}>
              Description
            </TableHead>
            <TableHead className={compact ? "text-right" : undefined}>
              Duration
            </TableHead>
            <TableHead className={compact ? "text-right" : undefined}>
              Date
            </TableHead>
            <TableHead className="w-10">
              <span className="sr-only">Actions</span>
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {entries.map((entry) => {
            const user = entry.userId ? users[entry.userId] : undefined;
            return (
              <TableRow key={entry.id} data-testid="timetracking-entry">
                <TableCell className={cn("whitespace-nowrap", cellClass)}>
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
                <TableCell
                  className={cn(
                    "min-w-0 text-left whitespace-normal",
                    compact ? "w-full" : "max-w-md",
                    cellClass
                  )}
                  data-testid="timetracking-entry-note"
                >
                  {entry.note || "No description"}
                </TableCell>
                <TableCell
                  className={cn(
                    "text-right font-medium whitespace-nowrap tabular-nums",
                    cellClass
                  )}
                >
                  {formatElapsed(entry.seconds)}
                </TableCell>
                <TableCell
                  className={cn(
                    "text-muted-foreground whitespace-nowrap",
                    compact && "text-right",
                    cellClass
                  )}
                >
                  {formatDate(entry.createdAt)}
                </TableCell>
                <TableCell className={cn("text-right whitespace-nowrap", cellClass)}>
                  <EntryMenu
                    compact={compact}
                    onEdit={() => setEdit(entry)}
                    onMove={() => setMove(entry)}
                    onDelete={() => setRemove(entry)}
                  />
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
      <EditEntryDialog
        entry={edit}
        pending={pending}
        onOpenChange={(open) => {
          if (!open) {
            setEdit(null);
          }
        }}
        onSave={async (seconds, note) => {
          if (!edit) {
            return;
          }
          setPending(true);
          try {
            await elemo.api.graph.nodes.update(edit.id, { seconds, note });
            setEdit(null);
            onChanged();
            notifySuccess("Time entry updated", "The duration and description were saved");
          } catch (cause) {
            notifyError("Failed to update time entry", cause);
          } finally {
            setPending(false);
          }
        }}
      />
      <MoveEntryDialog
        elemo={elemo}
        entry={move}
        projectId={projectId}
        pending={pending}
        onOpenChange={(open) => {
          if (!open) {
            setMove(null);
          }
        }}
        onMove={async (issueId) => {
          if (!move) {
            return;
          }
          setPending(true);
          try {
            await elemo.api.graph.nodes.move(move.id, {
              parentId: issueId,
              parentType: "Issue",
            });
            setMove(null);
            onChanged();
            notifySuccess(
              "Time entry moved",
              "The log now belongs to the selected work item"
            );
          } catch (cause) {
            notifyError("Failed to move time entry", cause);
          } finally {
            setPending(false);
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
        title="Delete time entry"
        description="This time log will be removed from the work item."
        isPending={pending}
        onConfirm={() => {
          if (!remove) {
            return;
          }
          setPending(true);
          void elemo.api.graph.nodes
            .delete(remove.id)
            .then(() => {
              setRemove(null);
              onChanged();
              notifySuccess("Time entry deleted", "The time log was removed");
            })
            .catch((cause: unknown) => {
              notifyError("Failed to delete time entry", cause);
            })
            .finally(() => setPending(false));
        }}
      />
    </>
  );
}

function EntryMenu({
  compact = false,
  onEdit,
  onMove,
  onDelete,
}: {
  compact?: boolean;
  onEdit: () => void;
  onMove: () => void;
  onDelete: () => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            type="button"
            size={compact ? "icon-xs" : "icon-sm"}
            variant="ghost"
            aria-label="Time entry actions"
          />
        }
      >
        <MoreHorizontal className={compact ? "size-3.5" : "size-4"} />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={onEdit}>
          <Pencil className="size-3.5" />
          Edit
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onMove}>
          <Move className="size-3.5" />
          Move
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

export function LogTimeDialog({
  open,
  pending,
  onOpenChange,
  onSave,
}: {
  open: boolean;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (seconds: number, note: string) => Promise<void>;
}) {
  const [hours, setHours] = useState(0);
  const [minutes, setMinutes] = useState(0);
  const [note, setNote] = useState("");

  useEffect(() => {
    if (open) {
      setHours(0);
      setMinutes(0);
      setNote("");
    }
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Log time</DialogTitle>
          <DialogDescription>
            Add a manual time entry on this work item.
          </DialogDescription>
        </DialogHeader>
        <DurationFields
          hours={hours}
          minutes={minutes}
          onHours={setHours}
          onMinutes={setMinutes}
        />
        <div className="space-y-2">
          <Label htmlFor="timetracking-log-note">Description</Label>
          <Textarea
            id="timetracking-log-note"
            data-testid="timetracking-log-note"
            value={note}
            onChange={(event) => setNote(event.target.value)}
            placeholder="Optional description"
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            disabled={pending}
            onClick={() =>
              void onSave(secondsFromHoursMinutes(hours, minutes), note.trim())
            }
          >
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function EditEntryDialog({
  entry,
  pending,
  onOpenChange,
  onSave,
}: {
  entry: TimeEntryView | null;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (seconds: number, note: string) => Promise<void>;
}) {
  const split = hoursMinutesFromSeconds(entry?.seconds ?? 0);
  const [hours, setHours] = useState(split.hours);
  const [minutes, setMinutes] = useState(split.minutes);
  const [note, setNote] = useState(entry?.note ?? "");

  useEffect(() => {
    const next = hoursMinutesFromSeconds(entry?.seconds ?? 0);
    setHours(next.hours);
    setMinutes(next.minutes);
    setNote(entry?.note ?? "");
  }, [entry]);

  return (
    <Dialog open={entry !== null} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit time entry</DialogTitle>
          <DialogDescription>
            Update the duration or description.
          </DialogDescription>
        </DialogHeader>
        <DurationFields
          hours={hours}
          minutes={minutes}
          onHours={setHours}
          onMinutes={setMinutes}
        />
        <div className="space-y-2">
          <Label htmlFor="timetracking-edit-note">Description</Label>
          <Textarea
            id="timetracking-edit-note"
            data-testid="timetracking-edit-note"
            value={note}
            onChange={(event) => setNote(event.target.value)}
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            disabled={pending}
            onClick={() =>
              void onSave(secondsFromHoursMinutes(hours, minutes), note.trim())
            }
          >
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function MoveEntryDialog({
  elemo,
  entry,
  projectId,
  pending,
  onOpenChange,
  onMove,
}: {
  elemo: ElemoPluginAPI;
  entry: TimeEntryView | null;
  projectId?: string;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  onMove: (issueId: string) => Promise<void>;
}) {
  const [issueId, setIssueId] = useState("");
  const [options, setOptions] = useState<{ value: string; title: string }[]>(
    []
  );

  useEffect(() => {
    if (!entry) {
      return;
    }
    setIssueId("");
    let cancelled = false;
    async function load() {
      let listProjectId = projectId;
      if (!listProjectId && entry?.parentId) {
        try {
          const issue = await elemo.api.issues.get(entry.parentId);
          listProjectId = issue.projectId;
        } catch {
          listProjectId = projectId;
        }
      }
      if (!listProjectId) {
        return;
      }
      try {
        const issues = await elemo.api.issues.list({
          projectId: listProjectId,
        });
        if (!cancelled) {
          setOptions(
            issues.map((issue) => ({
              value: issue.id,
              title: `${issue.key} ${issue.title}`,
            }))
          );
        }
      } catch {
        if (!cancelled) {
          setOptions([]);
        }
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [elemo, entry, projectId]);

  const filtered = useMemo(
    () => options.filter((option) => option.value !== entry?.parentId),
    [options, entry]
  );

  return (
    <Dialog open={entry !== null} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Move time entry</DialogTitle>
          <DialogDescription>
            Choose another work item in this project.
          </DialogDescription>
        </DialogHeader>
        <SearchableEntitySelect
          options={filtered}
          value={issueId || undefined}
          placeholder="Select a work item"
          searchPlaceholder="Search work items…"
          onValueChange={setIssueId}
        />
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            disabled={pending || !issueId || issueId === entry?.parentId}
            onClick={() => void onMove(issueId)}
          >
            Move
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function DurationFields({
  hours,
  minutes,
  onHours,
  onMinutes,
}: {
  hours: number;
  minutes: number;
  onHours: (value: number) => void;
  onMinutes: (value: number) => void;
}) {
  return (
    <div className="grid grid-cols-2 gap-3">
      <div className="space-y-2">
        <Label htmlFor="timetracking-hours">Hours</Label>
        <Input
          id="timetracking-hours"
          data-testid="timetracking-hours"
          type="number"
          min={0}
          value={hours}
          onChange={(event) => onHours(Number(event.target.value) || 0)}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="timetracking-minutes">Minutes</Label>
        <Input
          id="timetracking-minutes"
          data-testid="timetracking-minutes"
          type="number"
          min={0}
          max={59}
          value={minutes}
          onChange={(event) => onMinutes(Number(event.target.value) || 0)}
        />
      </div>
    </div>
  );
}
