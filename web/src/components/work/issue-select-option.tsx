import { KindRibbon, issueKindLabels } from "./kind-ribbon";
import { PriorityRibbon, issuePriorityLabels } from "./priority-ribbon";

import type { EntitySelectOption } from "@/components/ui/entity-select";
import { StatusIndicator } from "@/components/ui/status-indicator";
import type { PartialIssue } from "@/lib/api/types";
import { issueStatusLabels } from "@/lib/work/issue-adapter";

type IssueSelectFields = Pick<
  PartialIssue,
  "id" | "key" | "title" | "kind" | "status" | "priority"
>;

export function issueSelectSearchText(
  issue: Pick<PartialIssue, "kind" | "status" | "priority">
): string {
  return [
    issue.kind,
    issueKindLabels[issue.kind],
    issue.status,
    issueStatusLabels[issue.status],
    issue.priority,
    issuePriorityLabels[issue.priority],
  ]
    .filter(Boolean)
    .join(" ");
}

export function IssueSelectDetails({
  issue,
}: {
  issue: Pick<PartialIssue, "kind" | "status" | "priority">;
}) {
  return (
    <span className="flex items-center gap-3">
      <KindRibbon
        kind={issue.kind}
        className="gap-1.5"
        labelClassName="text-xs"
      />
      <StatusIndicator
        status={issue.status}
        label={issueStatusLabels[issue.status]}
        className="gap-1.5"
        labelClassName="text-xs"
      />
      <PriorityRibbon
        priority={issue.priority}
        className="gap-1.5"
        labelClassName="text-xs"
      />
    </span>
  );
}

export function issueToSelectOption(
  issue: IssueSelectFields
): EntitySelectOption {
  return {
    value: issue.id,
    title: `${issue.key} ${issue.title}`,
    searchText: issueSelectSearchText(issue),
    details: <IssueSelectDetails issue={issue} />,
  };
}
