import { useQuery } from "@tanstack/react-query";
import { useState } from "react";

import { issueToSelectOption } from "./issue-select-option";
import { KindRibbon, issueKindLabels, issueKinds } from "./kind-ribbon";
import { PriorityRibbon, issuePriorityLabels } from "./priority-ribbon";
import { calendarDateToUtcNoonIso, utcIsoToCalendarDate } from "./utils";

import { DatePicker } from "@/components/ui/date-picker";
import {
  EntityMultiSelect,
  SearchableEntitySelect,
} from "@/components/ui/entity-select";
import type { EntitySelectOption } from "@/components/ui/entity-select";
import {
  PropertyList,
  propertyControlClassName,
} from "@/components/ui/property-list";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { StatusIndicator } from "@/components/ui/status-indicator";
import { collectListedPage, collectedQueryKey } from "@/lib/api/cursor-pages";
import {
  v1LabelsGetOptions,
  v1ProjectsIssuesGetOptions,
} from "@/lib/api/query-options";
import { projectIdPath } from "@/lib/api/refs";
import { v1ProjectsIssuesGet } from "@/lib/api/sdk";
import type {
  Issue,
  IssueKind,
  IssuePatch,
  IssuePriority,
  IssueResolution,
  IssueStatus,
  Label,
  OrganizationMember,
  PartialLabel,
  PartialUser,
} from "@/lib/api/types";
import { cn, getInitials } from "@/lib/utils";
import { issueStatusLabels } from "@/lib/work/issue-adapter";
import {
  assignmentIdsEqual,
  normalizeAssignmentIds,
} from "@/lib/work/issue-edit";
import { ISSUE_RELATIONS_PAGE_SIZE } from "@/lib/work/issue-relations";
import {
  issueResolutionLabels,
  issueResolutions,
} from "@/lib/work/issue-resolution";
import { mergeWorkLabels } from "@/lib/work/resolve-work-labels";
import {
  mergeWorkPeople,
  personDisplayName,
} from "@/lib/work/resolve-work-people";
import { useOrganizationMembersForNamespace } from "@/lib/work/use-organization-members-for-namespace";

const issueStatuses: readonly IssueStatus[] = [
  "open",
  "in progress",
  "blocked",
  "review",
  "done",
  "closed",
];

const issuePriorities: readonly IssuePriority[] = [
  "highest",
  "high",
  "normal",
  "low",
  "lowest",
];

const UNSET_PARENT_VALUE = "__unset__";

export function IssueParentSelect({
  issue,
  disabled = false,
  onPatch,
}: {
  issue: Issue;
  disabled?: boolean;
  onPatch: (patch: IssuePatch, description?: string) => Promise<void>;
}) {
  const project = issue.project;
  const listOptions = v1ProjectsIssuesGetOptions({
    path: projectIdPath(project?.id ?? ""),
    query: { page_size: ISSUE_RELATIONS_PAGE_SIZE },
  });
  const { data: issuesPage } = useQuery({
    staleTime: listOptions.staleTime,
    gcTime: listOptions.gcTime,
    queryKey: collectedQueryKey(listOptions.queryKey),
    enabled: Boolean(project?.id),
    queryFn: async ({ signal }) =>
      collectListedPage(async (pageToken) => {
        const { data } = await v1ProjectsIssuesGet({
          path: projectIdPath(project?.id ?? ""),
          query: {
            page_size: ISSUE_RELATIONS_PAGE_SIZE,
            ...(pageToken ? { page_token: pageToken } : {}),
          },
          signal,
          throwOnError: true,
        });
        return data;
      }),
  });
  const catalogParents = (issuesPage?.items ?? []).filter(
    (candidate) => candidate.id !== issue.id
  );
  const parentCandidates =
    issue.parent &&
    !catalogParents.some((candidate) => candidate.id === issue.parent?.id)
      ? [issue.parent, ...catalogParents]
      : catalogParents;
  const parentOptions: EntitySelectOption[] = [
    {
      value: UNSET_PARENT_VALUE,
      title: "—",
      searchText: "none unset",
    },
    ...parentCandidates.map(issueToSelectOption),
  ];

  return (
    <SearchableEntitySelect
      size="sm"
      options={parentOptions}
      value={issue.parent?.id ?? UNSET_PARENT_VALUE}
      disabled={disabled || !project?.id}
      placeholder="—"
      searchPlaceholder="Search issues…"
      emptyMessage="No issues found."
      triggerClassName={propertyControlClassName}
      aria-label="Parent"
      onValueChange={(next) => {
        if (next === UNSET_PARENT_VALUE) {
          if (issue.parent) {
            void onPatch({ parent: null }, "Parent cleared");
          }
          return;
        }
        if (!next || next === issue.parent?.id) {
          return;
        }
        const selected = parentCandidates.find(
          (candidate) => candidate.id === next
        );
        void onPatch(
          { parent: next },
          selected ? `Parent set to ${selected.key}` : "Parent updated"
        );
      }}
    />
  );
}

function memberToOption(
  member: OrganizationMember | PartialUser
): EntitySelectOption {
  return {
    value: member.id,
    title: personDisplayName(member),
    avatarSrc: member.picture,
    avatarFallback: getInitials(member.first_name, member.last_name),
  };
}

function buildPersonOptions(
  selected: readonly PartialUser[],
  catalog?: readonly OrganizationMember[]
): EntitySelectOption[] {
  return mergeWorkPeople(selected, catalog).map((person) => {
    const selectedUser = selected.find((user) => user.id === person.id);
    const catalogUser = catalog?.find((user) => user.id === person.id);
    const source = catalogUser ?? selectedUser;
    if (source) {
      return memberToOption(source);
    }
    return {
      value: person.id,
      title: person.name,
      avatarSrc: person.picture,
      avatarFallback: person.name.slice(0, 2).toUpperCase(),
    };
  });
}

function labelToOption(
  label: Pick<Label, "id" | "name"> | PartialLabel
): EntitySelectOption {
  return {
    value: label.id,
    title: label.name,
  };
}

function buildLabelOptions(
  selected: readonly Pick<Label, "id" | "name">[],
  catalog?: readonly Pick<Label, "id" | "name">[]
): EntitySelectOption[] {
  return mergeWorkLabels(selected, catalog).map(labelToOption);
}

interface IssueDetailsPropertiesProps {
  issue: Issue;
  namespaceId: string;
  disabled?: boolean;
  onPatch: (patch: IssuePatch, description?: string) => Promise<void>;
}

export function IssueDetailsProperties({
  issue,
  namespaceId,
  disabled = false,
  onPatch,
}: IssueDetailsPropertiesProps) {
  const [loadMemberCatalog, setLoadMemberCatalog] = useState(false);
  const { members } = useOrganizationMembersForNamespace(namespaceId, {
    enabled: loadMemberCatalog,
  });

  const [loadLabelCatalog, setLoadLabelCatalog] = useState(false);
  const { data: labelsPage } = useQuery({
    ...v1LabelsGetOptions(),
    enabled: loadLabelCatalog,
  });

  const assigneeIds = normalizeAssignmentIds(
    issue.assignees.map((user) => user.id)
  );
  const reviewerIds = normalizeAssignmentIds(
    issue.reviewers.map((user) => user.id)
  );
  const labelIds = normalizeAssignmentIds(
    issue.labels.map((label) => label.id)
  );
  const assigneeOptions = buildPersonOptions(issue.assignees, members);
  const reviewerOptions = buildPersonOptions(issue.reviewers, members);
  const labelOptions = buildLabelOptions(issue.labels, labelsPage?.items);
  const dueDate = issue.due_date ? utcIsoToCalendarDate(issue.due_date) : null;
  const startDate = issue.start_date
    ? utcIsoToCalendarDate(issue.start_date)
    : null;

  return (
    <PropertyList
      items={[
        {
          label: "Kind",
          value: (
            <Select
              value={issue.kind}
              onValueChange={(value) => {
                if (!value || value === issue.kind) {
                  return;
                }
                void onPatch(
                  { kind: value as IssueKind },
                  `Kind set to ${issueKindLabels[value as IssueKind]}`
                );
              }}
              disabled={disabled}
              items={issueKindLabels}
            >
              <SelectTrigger
                size="sm"
                className={propertyControlClassName}
                aria-label="Kind"
              >
                <SelectValue>
                  <KindRibbon kind={issue.kind} />
                </SelectValue>
              </SelectTrigger>
              <SelectContent align="start" alignItemWithTrigger={false}>
                {issueKinds.map((kind) => (
                  <SelectItem key={kind} value={kind}>
                    <KindRibbon kind={kind} />
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          ),
        },
        {
          label: "Status",
          value: (
            <Select
              value={issue.status}
              onValueChange={(value) => {
                if (!value || value === issue.status) {
                  return;
                }
                void onPatch(
                  { status: value as IssueStatus },
                  `Status set to ${issueStatusLabels[value as IssueStatus]}`
                );
              }}
              disabled={disabled}
              items={issueStatusLabels}
            >
              <SelectTrigger
                size="sm"
                className={propertyControlClassName}
                aria-label="Status"
              >
                <SelectValue>
                  <StatusIndicator
                    status={issue.status}
                    label={issueStatusLabels[issue.status]}
                  />
                </SelectValue>
              </SelectTrigger>
              <SelectContent align="start" alignItemWithTrigger={false}>
                {issueStatuses.map((status) => (
                  <SelectItem key={status} value={status}>
                    <StatusIndicator
                      status={status}
                      label={issueStatusLabels[status]}
                    />
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          ),
        },
        {
          label: "Resolution",
          value: (
            <Select
              value={issue.resolution}
              onValueChange={(value) => {
                if (!value || value === issue.resolution) {
                  return;
                }
                void onPatch(
                  { resolution: value as IssueResolution },
                  `Resolution set to ${issueResolutionLabels[value as IssueResolution]}`
                );
              }}
              disabled={disabled}
              items={issueResolutionLabels}
            >
              <SelectTrigger
                size="sm"
                className={propertyControlClassName}
                aria-label="Resolution"
              >
                <SelectValue>
                  {issueResolutionLabels[issue.resolution]}
                </SelectValue>
              </SelectTrigger>
              <SelectContent align="start" alignItemWithTrigger={false}>
                {issueResolutions.map((resolution) => (
                  <SelectItem key={resolution} value={resolution}>
                    {issueResolutionLabels[resolution]}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          ),
        },
        {
          label: "Priority",
          value: (
            <Select
              value={issue.priority}
              onValueChange={(value) => {
                if (!value || value === issue.priority) {
                  return;
                }
                void onPatch(
                  { priority: value as IssuePriority },
                  `Priority set to ${issuePriorityLabels[value as IssuePriority]}`
                );
              }}
              disabled={disabled}
              items={issuePriorityLabels}
            >
              <SelectTrigger
                size="sm"
                className={propertyControlClassName}
                aria-label="Priority"
              >
                <SelectValue>
                  <PriorityRibbon priority={issue.priority} />
                </SelectValue>
              </SelectTrigger>
              <SelectContent align="start" alignItemWithTrigger={false}>
                {issuePriorities.map((priority) => (
                  <SelectItem key={priority} value={priority}>
                    <PriorityRibbon priority={priority} />
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          ),
        },
        {
          label: "Assignees",
          value: (
            <EntityMultiSelect
              size="sm"
              options={assigneeOptions}
              value={assigneeIds}
              disabled={disabled}
              placeholder="Unassigned"
              searchPlaceholder="Search people…"
              emptyMessage="No people found."
              triggerClassName={propertyControlClassName}
              aria-label="Assignees"
              onOpenChange={(open) => {
                if (open) {
                  setLoadMemberCatalog(true);
                }
              }}
              onValueChange={(next) => {
                const assignees = normalizeAssignmentIds(next);
                if (assignmentIdsEqual(assignees, assigneeIds)) {
                  return;
                }
                void onPatch(
                  { assignees },
                  assignees.length === 0
                    ? "Assignees cleared"
                    : "Assignees updated"
                );
              }}
            />
          ),
        },
        {
          label: "Reviewers",
          value: (
            <EntityMultiSelect
              size="sm"
              options={reviewerOptions}
              value={reviewerIds}
              disabled={disabled}
              placeholder="—"
              searchPlaceholder="Search people…"
              emptyMessage="No people found."
              triggerClassName={propertyControlClassName}
              aria-label="Reviewers"
              onOpenChange={(open) => {
                if (open) {
                  setLoadMemberCatalog(true);
                }
              }}
              onValueChange={(next) => {
                const reviewers = normalizeAssignmentIds(next);
                if (assignmentIdsEqual(reviewers, reviewerIds)) {
                  return;
                }
                void onPatch(
                  { reviewers },
                  reviewers.length === 0
                    ? "Reviewers cleared"
                    : "Reviewers updated"
                );
              }}
            />
          ),
        },
        {
          label: "Start date",
          value: (
            <DatePicker
              date={startDate}
              disabled={disabled}
              clearable
              placeholder="Start date"
              aria-label="Start date"
              clearAriaLabel="Clear start date"
              className={cn(propertyControlClassName, "justify-start")}
              onDateChange={(date) => {
                const next = date ? calendarDateToUtcNoonIso(date) : null;
                if (next === (issue.start_date ?? null)) {
                  return;
                }
                void onPatch(
                  { start_date: next },
                  date ? "Start date updated" : "Start date cleared"
                );
              }}
            />
          ),
        },
        {
          label: "Due date",
          value: (
            <DatePicker
              date={dueDate}
              disabled={disabled}
              clearable
              placeholder="Due date"
              aria-label="Due date"
              clearAriaLabel="Clear due date"
              className={cn(propertyControlClassName, "justify-start")}
              onDateChange={(date) => {
                const next = date ? calendarDateToUtcNoonIso(date) : null;
                if (next === (issue.due_date ?? null)) {
                  return;
                }
                void onPatch(
                  { due_date: next },
                  date ? "Due date updated" : "Due date cleared"
                );
              }}
            />
          ),
        },
        {
          label: "Labels",
          value: (
            <EntityMultiSelect
              size="sm"
              options={labelOptions}
              value={labelIds}
              disabled={disabled}
              placeholder="—"
              searchPlaceholder="Search labels…"
              emptyMessage="No labels found."
              triggerClassName={propertyControlClassName}
              aria-label="Labels"
              onOpenChange={(open) => {
                if (open) {
                  setLoadLabelCatalog(true);
                }
              }}
              onValueChange={(next) => {
                const labels = normalizeAssignmentIds(next);
                if (assignmentIdsEqual(labels, labelIds)) {
                  return;
                }
                void onPatch(
                  { labels },
                  labels.length === 0 ? "Labels cleared" : "Labels updated"
                );
              }}
            />
          ),
        },
      ]}
    />
  );
}
