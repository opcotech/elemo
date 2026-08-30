import {
  IssueMetadataProperties,
  IssueParentLink,
} from "./issue-metadata-properties";
import { IssueRelationsPreview } from "./issue-relations";
import { MarkdownContent } from "./markdown-content";
import { workItemPath, workItemUrl } from "./utils";
import { WorkItemDetailsReadonly } from "./work-item-details-readonly";

import { IssueCustomFields } from "@/components/custom-fields/issue-custom-fields";
import { ActivityFeed } from "@/components/shared/activity-feed";
import { MockDataAlert } from "@/components/shared/app-feedback";
import { EntityHeader } from "@/components/shared/entity-header";
import { RelationList } from "@/components/shared/relation-list";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { InternalLink } from "@/components/ui/internal-link";
import { Section, SectionAccordion } from "@/components/ui/section";
import { internalPath } from "@/lib/internal-url";
import { selectActivity, selectRelations } from "@/lib/mock-data";
import type { WorkItem } from "@/lib/work/model";
import { workItemPeople } from "@/lib/work/resolve-work-people";

export function WorkInspector({ item }: { item: WorkItem }) {
  const isApi = item.dataSource === "api";
  const relations = isApi
    ? []
    : selectRelations({
        entity: { id: item.id, type: "work-item" },
      });
  const activity = isApi
    ? []
    : selectActivity({
        entity: { id: item.id, type: "work-item" },
        limit: 4,
      });

  const assignees = workItemPeople(item.assignees, item.assigneeIds);
  const reviewers = workItemPeople(item.reviewers, item.reviewerIds);
  const namespaceId = item.namespace?.id ?? item.namespaceId;
  const namespaceLabel = item.namespace?.name ?? (namespaceId || "Unknown");
  const projectId = item.project?.id ?? item.projectId;
  const projectLabel = item.project
    ? item.project.name || item.project.key
    : projectId || "Unknown";

  return (
    <div className="space-y-6 p-4 pt-4 pr-12">
      <EntityHeader
        type="work-item"
        eyebrow={item.key}
        copyValue={workItemUrl(item)}
        copyLabel="Copy issue link"
        title={item.title}
        description={
          item.summary.trim() ? (
            <MarkdownContent markdown={item.summary} />
          ) : undefined
        }
        showIcon={false}
      />
      <Button
        className="w-full"
        render={<InternalLink to={internalPath(workItemPath(item))} />}
      >
        Open full page
      </Button>
      <Section data-section="issue-details">
        <WorkItemDetailsReadonly item={item} compact />
      </Section>
      {isApi && item.organizationSlug && item.namespaceSlug ? (
        <IssueRelationsPreview
          issueId={item.id}
          organizationSlug={item.organizationSlug}
          namespaceSlug={item.namespaceSlug}
        />
      ) : (
        <Section title="Relations" data-section="issue-relations">
          {relations.length > 0 ? (
            <RelationList
              relations={relations}
              entity={{ id: item.id, type: "work-item" }}
              limit={4}
            />
          ) : (
            <EmptyState
              compact
              title="No relations"
              description="Related work will appear here."
            />
          )}
        </Section>
      )}
      {isApi ? (
        <IssueCustomFields
          issueId={item.id}
          namespaceId={namespaceId}
          mode="readonly"
        />
      ) : null}
      <Section title="Metadata" data-section="issue-metadata">
        <IssueMetadataProperties
          compact
          organizationSlug={item.organizationSlug}
          namespaceSlug={item.namespaceSlug}
          namespaceLabel={namespaceLabel}
          projectKey={item.project?.key}
          projectLabel={projectLabel}
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
          reporterPeople={[...assignees, ...reviewers]}
          reporterDataSource={item.dataSource}
        />
      </Section>
      {isApi ? (
        <MockDataAlert title="Live issue detail">
          Core fields and relations are live. Activity stays unavailable until
          that API exists.
        </MockDataAlert>
      ) : (
        <MockDataAlert title="Illustrative work detail">
          Work properties, relationships, and activity shown here are
          illustrative examples.
        </MockDataAlert>
      )}
      <SectionAccordion title="Activity" value="activity">
        {activity.length > 0 ? (
          <ActivityFeed entries={activity} />
        ) : (
          <EmptyState
            compact
            title="No activity"
            description={
              isApi
                ? "Issue activity is not available from the API yet."
                : "Recent changes will appear here."
            }
          />
        )}
      </SectionAccordion>
    </div>
  );
}
