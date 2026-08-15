import { useOptimisticIssueFieldPatch } from "./use-optimistic-issue-field-patch";

import type { IssuePatch } from "@/lib/api/types";
import type { WorkItem } from "@/lib/mock-data";

export interface TimelineDateChange {
  readonly item: WorkItem;
  readonly startDate: string | null;
  readonly dueDate: string | null;
}

export function useTimelineIssueDates(projectId: string | undefined) {
  const { patchIssueFields, isPending } =
    useOptimisticIssueFieldPatch(projectId);

  const updateDates = async (change: TimelineDateChange) => {
    if (
      change.startDate === change.item.startDate &&
      change.dueDate === change.item.dueDate
    ) {
      return;
    }

    const patch: IssuePatch = {
      start_date: change.startDate,
      due_date: change.dueDate,
    };

    await patchIssueFields({
      item: change.item,
      patch,
      successDescription: "Schedule updated",
    });
  };

  return {
    updateDates,
    isPending,
  };
}
