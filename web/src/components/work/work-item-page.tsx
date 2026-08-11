import {
  ArrowLeftIcon,
  Code2Icon,
  FileQuestionIcon,
  MessageSquareIcon,
} from "lucide-react";
import { useEffect } from "react";

import { ContentWidth } from "@/components/layout/content-width";
import { ActivityFeed } from "@/components/shared/activity-feed";
import { AppEmptyState, MockDataAlert } from "@/components/shared/app-feedback";
import { EntityHeader, PageActions } from "@/components/shared/entity-header";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { PropertyList } from "@/components/shared/property-list";
import { RelationList } from "@/components/shared/relation-list";
import { Section } from "@/components/shared/section";
import { StatusIndicator } from "@/components/shared/status-indicator";
import { Button } from "@/components/ui/button";
import { InternalLink } from "@/components/ui/internal-link";
import { Textarea } from "@/components/ui/textarea";
import { internalPath } from "@/lib/internal-url";
import {
  getPerson,
  getWorkItem,
  selectActivity,
  selectRelations,
} from "@/lib/mock-data";
import { uiActions } from "@/lib/ui-store";

export function WorkItemPage({ workId }: { workId: string }) {
  const item = getWorkItem(workId);

  useEffect(() => {
    if (!item) return;
    uiActions.rememberRecentEntity({
      id: item.id,
      type: "work",
      label: `${item.key} ${item.title}`,
      href: internalPath(`/work/${item.id}`),
      namespaceId: item.namespaceId,
    });
  }, [item]);

  if (!item) {
    return (
      <ContentWidth width="entity">
        <AppEmptyState
          icon={<FileQuestionIcon />}
          title="Work item not found"
          description="This fixture-backed work item may not exist or may no longer be available."
          action={
            <Button variant="outline" render={<InternalLink to="/my-work" />}>
              <ArrowLeftIcon />
              Return to My Work
            </Button>
          }
        />
      </ContentWidth>
    );
  }

  const assignee = item.assigneeId ? getPerson(item.assigneeId) : undefined;
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
        copyValue={item.id}
        copyLabel="Copy work item ID"
        title={item.title}
        description={item.summary}
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
          <Section title="Description">
            <div className="bg-card min-h-40 rounded-xl border p-5">
              <p className="leading-7">{item.summary}</p>
              <p className="text-muted-foreground mt-4 leading-7">
                Acceptance details and a fuller description will appear here
                when editable work bodies are supported.
              </p>
            </div>
          </Section>
          <Section title="Linked documents">
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
              <AppEmptyState
                compact
                icon={<Code2Icon />}
                title="No linked documents"
                description="Specifications and decisions related to this work will appear here."
              />
            )}
          </Section>
          <Section title="Activity">
            <ActivityFeed entries={activity} />
          </Section>
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

        <aside className="space-y-8">
          <Section title="Details">
            <PropertyList
              items={[
                {
                  label: "Status",
                  value: <StatusIndicator status={item.status} />,
                },
                {
                  label: "Assignee",
                  value: assignee?.displayName ?? "Unassigned",
                },
                { label: "Priority", value: item.priority },
                {
                  label: "Target",
                  value: item.dueDate
                    ? new Intl.DateTimeFormat(undefined, {
                        month: "short",
                        day: "numeric",
                        year: "numeric",
                      }).format(new Date(item.dueDate))
                    : "Unscheduled",
                },
                {
                  label: "Estimate",
                  value:
                    item.estimatePoints === null
                      ? "Not estimated"
                      : `${item.estimatePoints} points`,
                },
                {
                  label: "Labels",
                  value: item.labelIds.join(", ") || "None",
                },
              ]}
            />
          </Section>
          <Section title="Relations">
            <RelationList
              relations={relations}
              entity={{ id: item.id, type: "work-item" }}
            />
          </Section>
        </aside>
      </div>
    </ContentWidth>
  );
}
