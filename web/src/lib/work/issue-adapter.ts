import type {
  Issue,
  IssueStatus,
  PartialIssue,
  PartialLabel,
  PartialUser,
} from "@/lib/api/types";
import type { WorkItem, WorkPerson, WorkStatus } from "@/lib/work/model";
import type { WorkLabel } from "@/lib/work/resolve-work-labels";

export interface IssueWorkContext {
  readonly namespaceId?: string;
  readonly organizationId?: string;
  readonly organizationSlug?: string;
  readonly namespaceSlug?: string;
  readonly projectId?: string | null;
}

export interface AccessibleNamespaceWorkContext {
  readonly id: string;
  readonly slug: string;
  readonly organizationId: string;
  readonly organizationSlug: string;
}

function issueWorkContextForNamespace(
  namespaces: ReadonlyMap<string, AccessibleNamespaceWorkContext>,
  namespaceId: string | undefined
): IssueWorkContext {
  const namespace = namespaceId ? namespaces.get(namespaceId) : undefined;
  if (!namespace) {
    return {};
  }
  return {
    namespaceId: namespace.id,
    namespaceSlug: namespace.slug,
    organizationId: namespace.organizationId,
    organizationSlug: namespace.organizationSlug,
  };
}

export const issueStatusLabels: Record<IssueStatus, string> = {
  open: "Backlog",
  "in progress": "In progress",
  blocked: "Blocked",
  review: "In review",
  done: "Done",
  closed: "Closed",
};

const issueStatusToWorkStatus: Record<IssueStatus, WorkStatus> = {
  open: "backlog",
  "in progress": "in progress",
  blocked: "blocked",
  review: "in review",
  done: "done",
  closed: "closed",
};

const workStatusToIssueStatus: Record<WorkStatus, IssueStatus> = {
  backlog: "open",
  "in progress": "in progress",
  "in review": "review",
  blocked: "blocked",
  done: "done",
  closed: "closed",
};

export function mapIssueStatus(status: IssueStatus): WorkStatus {
  return issueStatusToWorkStatus[status];
}

export function mapWorkStatusToIssueStatus(status: WorkStatus): IssueStatus {
  return workStatusToIssueStatus[status];
}

type IssueLike = Pick<
  Issue,
  "id" | "key" | "kind" | "title" | "status" | "priority" | "numeric_id"
> &
  Partial<
    Pick<
      Issue,
      | "assignees"
      | "reviewers"
      | "labels"
      | "due_date"
      | "start_date"
      | "description"
      | "created_at"
      | "updated_at"
      | "project"
      | "namespace"
      | "parent"
      | "resolution"
      | "links"
    >
  > & {
    reported_by?: PartialUser | null;
  };

/** Adapt an API Issue or PartialIssue into the Work UI model. */
export function issueToWorkItem(
  issue: IssueLike,
  context: IssueWorkContext = {}
): WorkItem {
  const assignees = issuePeople(issue.assignees);
  const reviewers = issuePeople(issue.reviewers);
  const labels = issueLabels(issue.labels);
  const namespaceId = issue.namespace?.id ?? context.namespaceId ?? "";
  const projectId = issue.project?.id ?? context.projectId ?? null;
  const namespaceSlug = issue.namespace?.slug ?? context.namespaceSlug;
  const organizationSlug = context.organizationSlug;
  const organizationId = context.organizationId;

  return {
    dataSource: "api",
    id: issue.id,
    key: issue.key,
    title: issue.title,
    summary: issue.description?.trim() ?? "",
    namespaceId,
    organizationId,
    organizationSlug,
    namespaceSlug,
    projectId,
    namespace: issue.namespace
      ? {
          id: issue.namespace.id,
          name: issue.namespace.name,
          slug: issue.namespace.slug,
          organizationId,
          organizationSlug,
        }
      : undefined,
    project: issue.project
      ? {
          id: issue.project.id,
          key: issue.project.key,
          name: issue.project.name,
        }
      : undefined,
    kind: issue.kind,
    resolution: issue.resolution,
    parent: issue.parent
      ? {
          id: issue.parent.id,
          key: issue.parent.key,
          title: issue.parent.title,
          namespaceId: issue.parent.namespace?.id,
          namespaceSlug: issue.parent.namespace?.slug,
          organizationSlug,
        }
      : issue.parent === null
        ? null
        : undefined,
    links: issue.links ?? [],
    status: mapIssueStatus(issue.status),
    priority: issue.priority,
    assigneeIds: assignees.map((person) => person.id),
    reviewerIds: reviewers.map((person) => person.id),
    assignees,
    reviewers,
    assigneeId: assignees[0]?.id ?? null,
    creatorId: issue.reported_by?.id ?? "",
    labelIds: labels.map((label) => label.id),
    labels,
    rank: issue.numeric_id,
    dueDate: issue.due_date ?? null,
    startDate: issue.start_date ?? null,
    createdAt: issue.created_at ?? new Date(0).toISOString(),
    updatedAt:
      issue.updated_at ?? issue.created_at ?? new Date(0).toISOString(),
  };
}

export function issuesToWorkItems(
  issues: readonly IssueLike[],
  context: IssueWorkContext = {}
): WorkItem[] {
  return issues.map((issue) => issueToWorkItem(issue, context));
}

/** Adapt assigned/cross-namespace issue lists using reachable namespace slugs. */
export function issuesToWorkItemsWithNamespaces(
  issues: readonly IssueLike[],
  namespaces: readonly AccessibleNamespaceWorkContext[]
): WorkItem[] {
  const byId = new Map(
    namespaces.map((namespace) => [namespace.id, namespace] as const)
  );
  return issues.map((issue) =>
    issueToWorkItem(
      issue,
      issueWorkContextForNamespace(byId, issue.namespace?.id)
    )
  );
}

export function partialIssuesToWorkItems(
  issues: readonly PartialIssue[],
  context: IssueWorkContext = {}
): WorkItem[] {
  return issuesToWorkItems(issues, context);
}

function issuePeople(users: readonly PartialUser[] | undefined): WorkPerson[] {
  return (users ?? []).map((user) => ({
    id: user.id,
    name: `${user.first_name} ${user.last_name}`.trim() || user.id,
    picture: user.picture,
  }));
}

function issueLabels(labels: readonly PartialLabel[] | undefined): WorkLabel[] {
  return (labels ?? []).map((label) => ({
    id: label.id,
    name: label.name,
  }));
}
