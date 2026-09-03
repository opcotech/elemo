import type { ElemoPluginAPI } from "@elemo/plugin-sdk";
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Textarea,
} from "@elemo/plugin-ui";
import { Play, Plus, Square } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import {
  LogTimeDialog,
  useIssueEntries,
  useResolvedUsers,
  type TimeEntryView,
} from "./entries";
import { formatElapsed, personName } from "./format";
import { notifyError, notifySuccess } from "./notify";

export function TimeTracker(props: Record<string, unknown>) {
  const elemo = props.elemo as ElemoPluginAPI;
  const issueId = String(props.issueId ?? "");
  const [running, setRunning] = useState(false);
  const [elapsed, setElapsed] = useState(0);
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState("");
  const [logOpen, setLogOpen] = useState(false);
  const { entries, reload } = useIssueEntries(elemo, issueId);
  const totals = useMemo(() => totalsByUser(entries), [entries]);
  const users = useResolvedUsers(
    elemo,
    totals.map((row) => row.userId)
  );

  useEffect(() => {
    if (!issueId) {
      return;
    }
    let cancelled = false;
    void elemo.api
      .invoke("timer.status", { issueId })
      .then((result) => {
        const status = result as { running?: boolean; elapsed?: number };
        if (!cancelled) {
          setRunning(Boolean(status.running));
          setElapsed(Number(status.elapsed ?? 0));
        }
      })
      .catch(() => {
        if (!cancelled) {
          setRunning(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [elemo, issueId]);

  useEffect(() => {
    if (!running) {
      return;
    }
    const timer = window.setInterval(() => {
      setElapsed((value) => value + 1);
    }, 1000);
    return () => window.clearInterval(timer);
  }, [running]);

  async function start() {
    setBusy(true);
    try {
      await elemo.api.invoke("timer.start", { issueId });
      setRunning(true);
      setElapsed(0);
      notifySuccess("Timer started", "The timer is running on this work item");
    } catch (cause) {
      notifyError("Failed to start timer", cause);
    } finally {
      setBusy(false);
    }
  }

  async function stop() {
    setBusy(true);
    try {
      await elemo.api.invoke("timer.stop", { issueId, note });
      setRunning(false);
      setElapsed(0);
      setNote("");
      reload();
      notifySuccess("Time logged", "The timer was stopped and saved");
    } catch (cause) {
      notifyError("Failed to stop timer", cause);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div data-testid="timetracking-sidebar" className="space-y-3">
      <Card size="sm">
        <CardHeader>
          <CardTitle>Time tracking</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-muted-foreground font-mono text-lg tabular-nums">
            {formatElapsed(elapsed)}
          </p>
          <div className="space-y-2">
            <Textarea
              data-testid="timetracking-description"
              value={note}
              onChange={(event) => setNote(event.target.value)}
              placeholder="Description (optional)"
              rows={2}
            />
          </div>
          <div className="flex flex-wrap gap-2">
            {running ? (
              <Button size="sm" disabled={busy} onClick={() => void stop()}>
                <Square className="size-3.5" />
                Stop
              </Button>
            ) : (
              <Button
                size="sm"
                disabled={busy || !issueId}
                onClick={() => void start()}
              >
                <Play className="size-3.5" />
                Start
              </Button>
            )}
            <Button
              size="sm"
              variant="outline"
              disabled={busy || !issueId || running}
              onClick={() => setLogOpen(true)}
            >
              <Plus className="size-3.5" />
              Log time
            </Button>
          </div>
        </CardContent>
        <LogTimeDialog
          open={logOpen}
          pending={busy}
          onOpenChange={setLogOpen}
          onSave={async (seconds, description) => {
            setBusy(true);
            try {
              await elemo.api.invoke("timer.log", {
                issueId,
                seconds,
                note: description,
              });
              setLogOpen(false);
              reload();
              notifySuccess("Time logged", "The time entry was saved");
            } catch (cause) {
              notifyError("Failed to log time", cause);
            } finally {
              setBusy(false);
            }
          }}
        />
      </Card>
      {totals.length > 0 ? (
        <Table
          data-testid="timetracking-sidebar-totals"
          containerClassName="border-0 bg-transparent rounded-none"
        >
          <TableHeader className="sr-only">
            <TableRow>
              <TableHead>Logged by</TableHead>
              <TableHead>Duration</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {totals.map((row) => {
              const user = row.userId ? users[row.userId] : undefined;
              return (
                <TableRow
                  key={row.userId || "unknown"}
                  className="hover:bg-transparent has-aria-expanded:bg-transparent data-[state=selected]:bg-transparent"
                >
                  <TableCell className="py-1 pl-4 pr-2">
                    {user ? personName(user) : "Unknown"}
                  </TableCell>
                  <TableCell className="py-1 pr-4 pl-2 text-right font-medium tabular-nums">
                    {formatElapsed(row.seconds)}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      ) : null}
    </div>
  );
}

function totalsByUser(
  entries: TimeEntryView[]
): Array<{ userId: string; seconds: number }> {
  const totals = new Map<string, number>();
  for (const entry of entries) {
    const id = entry.userId || "";
    totals.set(id, (totals.get(id) ?? 0) + entry.seconds);
  }
  return [...totals.entries()]
    .map(([userId, seconds]) => ({ userId, seconds }))
    .sort((a, b) => b.seconds - a.seconds);
}
