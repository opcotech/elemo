import { useSuspenseQuery } from "@tanstack/react-query";
import { Code2Icon, FileQuestionIcon, MessageSquareIcon } from "lucide-react";
import { useEffect, useState } from "react";

import { IssueDescriptionEditor } from "./issue-description-editor";
import {
  IssueDetailsProperties,
  IssueParentSelect,
} from "./issue-details-properties";
import { IssueDocumentsSection } from "./issue-documents";
import { IssueInlineTitle } from "./issue-inline-title";
import { IssueLinks } from "./issue-links";
import {
  IssueMetadataProperties,
  IssueParentLink,
} from "./issue-metadata-properties";
import { IssueRelations } from "./issue-relations";
import { MarkdownContent } from "./markdown-content";
import { useIssueUpdate } from "./use-issue-update";
import { workItemPath, workItemUrl } from "./utils";
import { WorkItemDetailsReadonly } from "./work-item-details-readonly";

import { ContentWidth } from "@/components/layout/content-width";
import { ActivityFeed } from "@/components/shared/activity-feed";
import { MockDataAlert } from "@/components/shared/app-feedback";
import { EntityHeader, PageActions } from "@/components/shared/entity-header";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { RelationList } from "@/components/shared/relation-list";
import { Button } from "@/components/ui/button";
import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { EmptyState } from "@/components/ui/empty-state";
import { Section, SectionAccordion } from "@/components/ui/section";
import { Textarea } from "@/components/ui/textarea";
import { useDeleteMutation } from "@/hooks/use-delete-mutation";
import { v1IssueDeleteMutation } from "@/lib/api/mutation-options";
import {
  v1IssueGetOptions,
  v1NamespacesIssuesKeyGetOptions,
  v1ProjectsIssuesGetOptions,
} from "@/lib/api/query-options";
import { namespaceRefPath, projectIdPath } from "@/lib/api/refs";
import type { Issue, Options, V1IssueDeleteData } from "@/lib/api/types";
import { internalPath } from "@/lib/internal-url";
import { selectActivity, selectRelations } from "@/lib/mock-data";
import { uiActions } from "@/lib/ui-store";
import { issueToWorkItem } from "@/lib/work/issue-adapter";
import type { WorkItem } from "@/lib/work/model";
import {
  partialUserToPerson,
  workItemPeople,
} from "@/lib/work/resolve-work-people";

function MockWorkItemPage({ item }: { item: WorkItem }) {
  const recentHref = internalPath(workItemPath(item));

  useEffect(() => {
    uiActions.rememberRecentEntity({
      id: item.id,
      type: "work-item",
      label: `${item.key} ${item.title}`,
      href: recentHref,
      namespaceId: item.namespaceId || undefined,
    });
  }, [item.id, item.key, item.title, item.namespaceId, recentHref]);

  const relations = selectRelations({
    entity: { id: item.id, type: "work-item" },
  });
  const activity = selectActivity({
    entity: { id: item.id, type: "work-item" },
  });
  const linkedDocuments = relations
    .flatMap((relation) => [relation.from, relation.to])
    .filter((ref) => ref.type === "document");

  return (
    <ContentWidth width="overview" className="space-y-7">
      <EntityHeader
        type="work-item"
        eyebrow={item.key}
        copyValue={workItemUrl(item)}
        copyLabel="Copy issue link"
        title={item.title}
        description={<MarkdownContent markdown={item.summary} />}
        showIcon={false}
        actions={
          <PageActions
            secondary={[
              {
                label: "View relationships",
                href: `/relations/work-item/${item.id}`,
              },
              { label: "Copy link" },
            ]}
          />
        }
      />
      <MockDataAlert title="Illustrative work item">
        This canonical page is powered by centralized read-only work,
        relationship, activity, and people fixtures. No unsupported mutation is
        presented as persisted.
      </MockDataAlert>

      <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <main className="space-y-8">
          <Section title="Description" data-section="issue-description">
            <div className="bg-card min-h-40 rounded-xl border p-5">
              <MarkdownContent
                markdown={item.summary}
                size="default"
                empty={
                  <p className="text-muted-foreground leading-7">
                    No description
                  </p>
                }
              />
            </div>
          </Section>
          <Section title="Links" data-section="issue-links">
            {(item.links ?? []).length > 0 ? (
              <ul className="space-y-1 text-sm">
                {(item.links ?? []).map((link) => (
                  <li key={link.url}>
                    <a
                      href={link.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary underline-offset-4 hover:underline"
                    >
                      {link.label}
                    </a>
                    <span className="text-muted-foreground"> {link.url}</span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-muted-foreground text-sm">No links</p>
            )}
          </Section>
          <Section title="Relations" data-section="issue-relations">
            {relations.length > 0 ? (
              <RelationList
                relations={relations}
                entity={{ id: item.id, type: "work-item" }}
              />
            ) : (
              <EmptyState
                compact
                icon={<FileQuestionIcon />}
                title="No relations"
                description="Related work and documents will appear here."
              />
            )}
          </Section>
          <SectionAccordion
            title="Linked documents"
            value="documents"
            defaultOpen={linkedDocuments.length > 0}
          >
            {linkedDocuments.length > 0 ? (
              <AppList>
                {linkedDocuments.map((document) => (
                  <EntityLink
                    key={document.id}
                    type="document"
                    href={`/documents/${document.id}`}
                    title={document.title}
                  />
                ))}
              </AppList>
            ) : (
              <EmptyState
                compact
                icon={<Code2Icon />}
                title="No linked documents"
                description="Specifications and decisions related to this work will appear here."
              />
            )}
          </SectionAccordion>
          <SectionAccordion title="Activity" value="activity">
            {activity.length > 0 ? (
              <ActivityFeed entries={activity} />
            ) : (
              <EmptyState
                compact
                icon={<MessageSquareIcon />}
                title="No activity yet"
                description="Activity for this fixture will appear here when present."
              />
            )}
          </SectionAccordion>
          <Section title="Comment">
            <div className="space-y-3 rounded-xl border p-4">
              <Textarea
                disabled
                placeholder="Comments are not available for this work item yet."
              />
              <Button disabled>
                <MessageSquareIcon />
                Send
              </Button>
            </div>
          </Section>
        </main>

        <aside className="space-y-4">
          <Section title="Details" data-section="issue-details">
            <WorkItemDetailsReadonly item={item} />
          </Section>
          <Section title="Metadata" data-section="issue-metadata">
            <IssueMetadataProperties
              organizationSlug={item.organizationSlug}
              namespaceSlug={item.namespaceSlug}
              namespaceId={item.namespaceId}
              namespaceLabel={item.namespace?.name ?? item.namespaceId}
              projectKey={item.project?.key}
              projectLabel={
                item.project?.name ||
                item.project?.key ||
                item.projectId ||
                "Unknown"
              }
              parent={
                <IssueParentLink
                  parent={item.parent}
                  organizationSlug={item.organizationSlug}
                  namespaceSlug={item.namespaceSlug}
                />
              }
              reportedById={item.creatorId}
              createdAt={item.createdAt}
              updatedAt={item.updatedAt}
              reporterPeople={[
                ...workItemPeople(item.assignees, item.assigneeIds),
                ...workItemPeople(item.reviewers, item.reviewerIds),
              ]}
              reporterDataSource="mock"
            />
          </Section>
        </aside>
      </div>
    </ContentWidth>
  );
}

function LiveWorkItemPage({
  organizationId,
  organizationSlug,
  namespaceId,
  namespaceSlug,
  issueKey,
  initialIssue,
}: {
  organizationId: string;
  organizationSlug: string;
  namespaceId: string;
  namespaceSlug: string;
  issueKey: string;
  initialIssue: Issue;
}) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const issueKeyPath = {
    ...namespaceRefPath(organizationId, namespaceId),
    key: issueKey,
  };
  const { data: issue } = useSuspenseQuery({
    ...v1NamespacesIssuesKeyGetOptions({
      path: issueKeyPath,
    }),
    initialData: initialIssue,
  });

  const project = issue.project;
  const item = issueToWorkItem(issue, {
    namespaceId,
    organizationId,
    organizationSlug,
    namespaceSlug,
  });

  const { updateIssue, isPending } = useIssueUpdate({
    organizationId,
    namespaceId,
    issueKey,
    issueId: issue.id,
    projectId: project?.id ?? null,
  });

  const queryKeysToInvalidate = [
    v1NamespacesIssuesKeyGetOptions({
      path: issueKeyPath,
    }).queryKey,
    v1IssueGetOptions({
      path: { id: issue.id },
    }).queryKey,
    ...(project
      ? [
          v1ProjectsIssuesGetOptions({
            path: projectIdPath(project.id),
            query: { page_size: 100 },
          }).queryKey,
        ]
      : []),
  ];

  const deleteMutation = useDeleteMutation<void, Options<V1IssueDeleteData>>({
    mutationOptions: v1IssueDeleteMutation(),
    successMessage: "Issue deleted",
    successDescription: `${issue.key} has been deleted`,
    errorMessagePrefix: "Failed to delete issue",
    queryKeysToInvalidate,
    navigateOnSuccess: (navigate) => {
      if (project) {
        return navigate({
          to: "/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey/work",
          params: {
            organizationSlug,
            namespaceSlug,
            projectKey: project.key,
          },
          search: {
            group: "status",
            sort: "rank:asc",
            display: "comfortable",
            layout: "board",
          },
        });
      }
      return navigate({
        to: "/organizations/$organizationSlug/namespaces/$namespaceSlug",
        params: { organizationSlug, namespaceSlug },
      });
    },
  });

  const recentHref = internalPath(
    workItemPath({
      organizationSlug,
      namespaceSlug,
      key: issue.key,
    })
  );

  useEffect(() => {
    uiActions.rememberRecentEntity({
      id: issue.id,
      type: "work-item",
      label: `${issue.key} ${issue.title}`,
      href: recentHref,
      namespaceId,
    });
  }, [issue.id, issue.key, issue.title, recentHref, namespaceId]);

  return (
    <ContentWidth width="overview" className="space-y-7">
      <EntityHeader
        type="work-item"
        eyebrow={issue.key}
        copyValue={workItemUrl(item)}
        copyLabel="Copy issue link"
        title={
          <IssueInlineTitle
            title={issue.title}
            disabled={isPending || deleteMutation.isPending}
            onCommit={async (title) => {
              await updateIssue({ title }, "Title updated");
            }}
          />
        }
        showIcon={false}
        actions={
          <PageActions
            secondary={[
              {
                label: "View relationships",
                href: `/relations/work-item/${issue.id}`,
              },
              { label: "Copy link" },
              {
                label: "Delete",
                variant: "destructive",
                onSelect: () => setDeleteDialogOpen(true),
              },
            ]}
          />
        }
      />
      <DeleteConfirmationDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title={`Delete ${issue.key}?`}
        description={`Are you sure you want to delete “${issue.title}”? This cannot be undone.`}
        consequences={[
          "The issue will be permanently removed",
          "Related activity and comments will no longer be available",
        ]}
        deleteButtonText="Delete issue"
        isPending={deleteMutation.isPending}
        onConfirm={() => {
          deleteMutation.mutate({ path: { id: issue.id } });
        }}
      />
      <MockDataAlert title="Live issue with illustrative extras">
        Core issue fields and relations are editable. Activity, comments, and
        people profiles remain unavailable or fixture-backed until those APIs
        exist.
      </MockDataAlert>

      <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <main className="space-y-8">
          <Section data-section="issue-description">
            <IssueDescriptionEditor
              description={issue.description}
              namespaceId={namespaceId}
              disabled={isPending || deleteMutation.isPending}
              onCommit={async (description) => {
                await updateIssue(
                  { description },
                  description ? "Description updated" : "Description cleared"
                );
              }}
            />
          </Section>
          <IssueLinks
            links={issue.links ?? []}
            disabled={isPending || deleteMutation.isPending}
            onPatch={updateIssue}
          />
          <IssueRelations
            issueId={issue.id}
            issueKey={issue.key}
            organizationId={organizationId}
            organizationSlug={organizationSlug}
            namespaceId={namespaceId}
            namespaceSlug={namespaceSlug}
            disabled={isPending || deleteMutation.isPending}
          />
          <IssueDocumentsSection
            issueId={issue.id}
            organizationId={organizationId}
            namespaceId={namespaceId}
            issueKey={issueKey}
            documentCount={issue.document_count}
            canCreate={!isPending && !deleteMutation.isPending}
          />
          <SectionAccordion title="Activity" value="activity">
            <EmptyState
              compact
              icon={<MessageSquareIcon />}
              title="No activity yet"
              description="Issue activity is not available from the API yet."
            />
          </SectionAccordion>
          <Section title="Comment">
            <div className="space-y-3 rounded-xl border p-4">
              <Textarea
                disabled
                placeholder="Comments are not available for this work item yet."
              />
              <Button disabled>
                <MessageSquareIcon />
                Send
              </Button>
            </div>
          </Section>
        </main>

        <aside className="space-y-4">
          <Section title="Details" data-section="issue-details">
            <IssueDetailsProperties
              issue={issue}
              namespaceId={namespaceId}
              disabled={isPending || deleteMutation.isPending}
              onPatch={updateIssue}
            />
          </Section>
          <Section title="Metadata" data-section="issue-metadata">
            <IssueMetadataProperties
              organizationSlug={organizationSlug}
              namespaceSlug={namespaceSlug}
              namespaceId={namespaceId}
              namespaceLabel={issue.namespace?.name ?? namespaceId}
              projectKey={project?.key}
              projectLabel={project?.name || project?.key || "Unknown"}
              parent={
                <IssueParentSelect
                  issue={issue}
                  disabled={isPending || deleteMutation.isPending}
                  onPatch={updateIssue}
                />
              }
              reportedById={issue.reported_by?.id}
              createdAt={issue.created_at}
              updatedAt={issue.updated_at}
              reporterPeople={[
                ...issue.assignees.map(partialUserToPerson),
                ...issue.reviewers.map(partialUserToPerson),
              ]}
            />
          </Section>
        </aside>
      </div>
    </ContentWidth>
  );
}

export function WorkItemPage({
  item,
  issue,
  organizationId,
  organizationSlug,
  namespaceId,
  namespaceSlug,
  issueKey,
}: {
  item: WorkItem;
  issue?: Issue;
  organizationId?: string;
  organizationSlug?: string;
  namespaceId?: string;
  namespaceSlug?: string;
  issueKey?: string;
}) {
  if (
    issue &&
    organizationId &&
    organizationSlug &&
    namespaceId &&
    namespaceSlug &&
    issueKey
  ) {
    return (
      <LiveWorkItemPage
        organizationId={organizationId}
        organizationSlug={organizationSlug}
        namespaceId={namespaceId}
        namespaceSlug={namespaceSlug}
        issueKey={issueKey}
        initialIssue={issue}
      />
    );
  }

  return <MockWorkItemPage item={item} />;
}
