import { ActivityIcon, UserIcon } from "lucide-react";

import { EmptyState } from "@/components/ui/empty-state";
import { getPerson } from "@/lib/mock-data";
import type { ActivityEntry, Person } from "@/lib/mock-data/types";

export function ActivityFeed({
  entries,
  people,
}: {
  entries: readonly ActivityEntry[];
  people?: readonly Person[];
}) {
  if (entries.length === 0) {
    return (
      <EmptyState
        compact
        icon={<ActivityIcon />}
        title="No recent activity"
        description="Changes and comments will appear here."
      />
    );
  }

  return (
    <ol className="space-y-1">
      {entries.map((entry) => {
        const actor =
          people?.find((person) => person.id === entry.actorId) ??
          getPerson(entry.actorId);
        return (
          <li
            key={entry.id}
            className="hover:bg-muted/50 flex gap-3 rounded-lg px-2 py-2.5"
          >
            <span className="bg-muted text-muted-foreground mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full">
              <UserIcon className="size-3.5" />
            </span>
            <div className="min-w-0 text-sm">
              <p>
                <span className="font-medium">
                  {actor?.displayName ?? "A teammate"}
                </span>{" "}
                <span className="text-muted-foreground">{entry.detail}</span>
              </p>
              <time className="text-muted-foreground text-xs">
                {new Intl.DateTimeFormat(undefined, {
                  month: "short",
                  day: "numeric",
                  hour: "numeric",
                  minute: "2-digit",
                }).format(new Date(entry.occurredAt))}
              </time>
            </div>
          </li>
        );
      })}
    </ol>
  );
}
