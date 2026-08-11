import { dateLabel } from "./utils";

import { ActivityFeed } from "@/components/shared/activity-feed";
import { MockDataAlert } from "@/components/shared/app-feedback";
import { EntityHeader } from "@/components/shared/entity-header";
import { PropertyList } from "@/components/shared/property-list";
import { RelationList } from "@/components/shared/relation-list";
import { Section } from "@/components/shared/section";
import { StatusIndicator } from "@/components/shared/status-indicator";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { InternalLink } from "@/components/ui/internal-link";
import { internalPath } from "@/lib/internal-url";
import { getPerson, selectActivity, selectRelations } from "@/lib/mock-data";
import type { WorkItem } from "@/lib/mock-data";

export function WorkInspector({ item }: { item: WorkItem }) {
  const relations = selectRelations({
    entity: { id: item.id, type: "work-item" },
  });
  const activity = selectActivity({
    entity: { id: item.id, type: "work-item" },
    limit: 4,
  });
  const assignee = item.assigneeId ? getPerson(item.assigneeId) : undefined;

  return (
    <div className="space-y-6 p-4 pt-4 pr-12">
      <EntityHeader
        type="work-item"
        eyebrow={item.key}
        copyValue={item.id}
        copyLabel="Copy work item ID"
        title={item.title}
        description={item.summary}
        showIcon={false}
      />
      <Button
        className="w-full"
        render={<InternalLink to={internalPath(`/work/${item.id}`)} />}
      >
        Open full page
      </Button>
      <PropertyList
        compact
        items={[
          { label: "Status", value: <StatusIndicator status={item.status} /> },
          {
            label: "Assignee",
            value: assignee?.displayName ?? "Unassigned",
          },
          { label: "Priority", value: item.priority },
          { label: "Target", value: dateLabel(item.dueDate) },
          {
            label: "Labels",
            value:
              item.labelIds.length > 0
                ? item.labelIds.map((label) => (
                    <Badge key={label} variant="secondary" className="mr-1">
                      {label}
                    </Badge>
                  ))
                : "None",
          },
        ]}
      />
      <MockDataAlert title="Illustrative work detail">
        Work properties, relationships, and activity shown here are illustrative
        examples.
      </MockDataAlert>
      <Section title="Relations">
        <RelationList
          relations={relations}
          entity={{ id: item.id, type: "work-item" }}
          limit={4}
        />
      </Section>
      <Section title="Activity">
        <ActivityFeed entries={activity} />
      </Section>
    </div>
  );
}
