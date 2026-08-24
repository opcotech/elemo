import { KindRibbon } from "./kind-ribbon";
import { PriorityRibbon } from "./priority-ribbon";
import { formatTargetDate } from "./utils";
import { WorkLabelBadges } from "./work-label-badges";

import { PersonAvatarStack } from "@/components/ui/person-avatar-stack";
import { PropertyList } from "@/components/ui/property-list";
import { StatusIndicator } from "@/components/ui/status-indicator";
import { issueStatusLabels } from "@/lib/work/issue-adapter";
import { issueResolutionLabels } from "@/lib/work/issue-resolution";
import type { WorkItem } from "@/lib/work/model";
import { workItemPeople } from "@/lib/work/resolve-work-people";

export function WorkItemDetailsReadonly({
  item,
  compact = false,
}: {
  item: WorkItem;
  compact?: boolean;
}) {
  const assignees = workItemPeople(item.assignees, item.assigneeIds);
  const reviewers = workItemPeople(item.reviewers, item.reviewerIds);
  const statusLabel =
    item.status === "backlog" ? issueStatusLabels.open : undefined;

  return (
    <PropertyList
      compact={compact}
      items={[
        {
          label: "Kind",
          value: item.kind ? <KindRibbon kind={item.kind} /> : "—",
        },
        {
          label: "Status",
          value: <StatusIndicator status={item.status} label={statusLabel} />,
        },
        {
          label: "Resolution",
          value: item.resolution
            ? issueResolutionLabels[item.resolution]
            : "None",
        },
        {
          label: "Priority",
          value: <PriorityRibbon priority={item.priority} />,
        },
        {
          label: "Assignees",
          value: (
            <PersonAvatarStack
              people={assignees}
              size="sm"
              showNames
              emptyLabel="Unassigned"
            />
          ),
        },
        {
          label: "Reviewers",
          value: (
            <PersonAvatarStack
              people={reviewers}
              size="sm"
              showNames
              emptyLabel="None"
            />
          ),
        },
        {
          label: "Start date",
          value: formatTargetDate(item.startDate ?? null),
        },
        {
          label: "Due date",
          value: formatTargetDate(item.dueDate ?? null),
        },
        {
          label: "Labels",
          value:
            item.labelIds.length > 0 ? (
              <WorkLabelBadges
                labelIds={item.labelIds}
                labels={item.labels}
                limit={item.labelIds.length}
                truncate={false}
              />
            ) : (
              "None"
            ),
        },
      ]}
    />
  );
}
