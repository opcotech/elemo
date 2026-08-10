import { useState } from "react";

import { dateLabel, paginate } from "./utils";

import { StatusIndicator } from "@/components/shared/status-indicator";
import { Button } from "@/components/ui/button";
import { InternalLink } from "@/components/ui/internal-link";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { internalPath } from "@/lib/internal-url";
import { getPerson } from "@/lib/mock-data";
import type { WorkItem } from "@/lib/mock-data";
import { cn } from "@/lib/utils";

const TABLE_PAGE_SIZE = 50;

export function WorkTable({
  items,
  compact,
  onSelect,
}: {
  items: readonly WorkItem[];
  compact: boolean;
  onSelect: (item: WorkItem) => void;
}) {
  const [requestedPage, setRequestedPage] = useState(1);
  const page = paginate(items, requestedPage, TABLE_PAGE_SIZE);

  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col overflow-hidden">
      <ScrollArea className="min-h-0 min-w-0 flex-1">
        <Table className="min-w-195">
          <TableHeader className="bg-background sticky top-0 z-10">
            <TableRow>
              <TableHead className="bg-background sticky left-0">Key</TableHead>
              <TableHead className="min-w-72">Title</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Assignee</TableHead>
              <TableHead>Priority</TableHead>
              <TableHead>Target</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {page.items.map((item) => (
              <TableRow key={item.id}>
                <TableCell
                  className={cn(
                    "bg-background sticky left-0 font-mono text-xs",
                    compact ? "py-2" : "py-3"
                  )}
                >
                  <InternalLink
                    to={internalPath(`/work/${item.id}`)}
                    className="hover:text-primary"
                  >
                    {item.key}
                  </InternalLink>
                </TableCell>
                <TableCell className={cn(compact ? "py-2" : "py-3")}>
                  <button
                    type="button"
                    className="hover:text-primary focus-visible:ring-ring rounded-sm text-left font-medium outline-none focus-visible:ring-2"
                    aria-label={`Inspect ${item.key}: ${item.title}`}
                    onClick={() => onSelect(item)}
                  >
                    {item.title}
                  </button>
                </TableCell>
                <TableCell className={cn(compact ? "py-2" : "py-3")}>
                  <StatusIndicator status={item.status} />
                </TableCell>
                <TableCell className={cn(compact ? "py-2" : "py-3")}>
                  {item.assigneeId
                    ? (getPerson(item.assigneeId)?.displayName ?? "Assigned")
                    : "Unassigned"}
                </TableCell>
                <TableCell
                  className={cn("capitalize", compact ? "py-2" : "py-3")}
                >
                  {item.priority}
                </TableCell>
                <TableCell className={cn(compact ? "py-2" : "py-3")}>
                  {dateLabel(item.dueDate)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <ScrollBar orientation="horizontal" />
      </ScrollArea>
      {page.pageCount > 1 && (
        <nav
          aria-label="Work table pages"
          className="flex shrink-0 items-center justify-end gap-2 border-t px-3 py-2"
        >
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={page.page === 1}
            onClick={() => setRequestedPage(page.page - 1)}
          >
            Previous
          </Button>
          <span
            aria-live="polite"
            className="text-muted-foreground text-xs tabular-nums"
          >
            Page {page.page} of {page.pageCount}
          </span>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={page.page === page.pageCount}
            onClick={() => setRequestedPage(page.page + 1)}
          >
            Next
          </Button>
        </nav>
      )}
    </div>
  );
}
