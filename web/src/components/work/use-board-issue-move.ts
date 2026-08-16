import { useOptimisticIssueFieldPatch } from "./use-optimistic-issue-field-patch";

import type { IssuePatch } from "@/lib/api/types";
import type { WorkItem, WorkStatus } from "@/lib/mock-data";
import { mapWorkStatusToIssueStatus } from "@/lib/work/issue-adapter";

export type BoardMoveGroup = "status" | "priority";

export interface BoardItemMove {
  readonly item: WorkItem;
  readonly group: BoardMoveGroup;
  readonly from: string;
  readonly to: string;
}

function patchFromMove(move: BoardItemMove): IssuePatch | null {
  if (move.group === "status") {
    return { status: mapWorkStatusToIssueStatus(move.to as WorkStatus) };
  }
  if (move.group === "priority") {
    return {
      priority: move.to as IssuePatch["priority"],
    };
  }
  return null;
}

function successDescription(move: BoardItemMove): string {
  if (move.group === "status") {
    return `Status set to ${move.to}`;
  }
  return `Priority set to ${move.to}`;
}

export function useBoardIssueMove(projectId: string | undefined) {
  const { patchIssueFields, isPending } =
    useOptimisticIssueFieldPatch(projectId);

  const moveIssue = async (move: BoardItemMove) => {
    if (move.from === move.to) {
      return;
    }

    const patch = patchFromMove(move);
    if (!patch) {
      return;
    }

    await patchIssueFields({
      item: move.item,
      patch,
      successDescription: successDescription(move),
    });
  };

  return {
    moveIssue,
    isPending,
  };
}
