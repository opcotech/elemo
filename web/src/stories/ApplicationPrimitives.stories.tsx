import type { Meta, StoryObj } from "@storybook/react-vite";
import { BellRingIcon, Link2Icon, PlusIcon } from "lucide-react";

import { withRouter } from "../../.storybook/with-router";

import { ActivityFeed } from "@/components/shared/activity-feed";
import { AppEmptyState, MockDataAlert } from "@/components/shared/app-feedback";
import { CreateButton } from "@/components/shared/create-button";
import { EntityHeader } from "@/components/shared/entity-header";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { PropertyList } from "@/components/shared/property-list";
import { RelationList } from "@/components/shared/relation-list";
import { Section } from "@/components/shared/section";
import { StatusIndicator } from "@/components/shared/status-indicator";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EntitySelect } from "@/components/ui/entity-select";
import {
  mockPeople,
  mockWorkItems,
  selectActivity,
  selectRelations,
} from "@/lib/mock-data";

const selectedWork = mockWorkItems[0];
const selectedEntity = {
  id: selectedWork.id,
  type: "work-item" as const,
};

const meta: Meta = {
  title: "Elemo/Application Primitives",
  decorators: [withRouter],
  parameters: {
    a11y: { test: "error" },
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "Reusable entity, context, empty-state, mock-boundary, relationship, property, and activity primitives used by Elemo operational pages.",
      },
    },
  },
  tags: ["autodocs", "verification"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Populated: Story = {
  render: () => (
    <div className="mx-auto max-w-5xl space-y-8 p-6">
      <EntityHeader
        type="project"
        eyebrow="Project · Product namespace"
        title="Elemo web experience"
        description="A representative populated entity header with progressive actions and operational context."
        meta={
          <div className="flex items-center gap-3">
            <StatusIndicator status="in-progress" />
            <Badge variant="secondary">6 contributors</Badge>
          </div>
        }
        actions={
          <>
            <Button variant="outline">Share</Button>
            <CreateButton>Create work</CreateButton>
          </>
        }
      />

      <MockDataAlert>
        The work, relationship, and activity regions in this story use
        centralized illustrative fixtures.
      </MockDataAlert>

      <div className="grid gap-8 lg:grid-cols-2">
        <Section title="Entity links" description="Canonical destinations">
          <AppList>
            {mockWorkItems.slice(0, 3).map((item) => (
              <EntityLink
                key={item.id}
                href={`/work/${item.id}`}
                type="work-item"
                title={`${item.key} ${item.title}`}
                subtitle={item.summary}
              />
            ))}
          </AppList>
        </Section>

        <Section title="Properties">
          <PropertyList
            items={[
              {
                label: "Status",
                value: <StatusIndicator status={selectedWork.status} />,
              },
              { label: "Priority", value: selectedWork.priority },
              { label: "Target", value: selectedWork.dueDate ?? "Unscheduled" },
              {
                label: "Labels",
                value: selectedWork.labelIds.map((label) => (
                  <Badge key={label} variant="secondary" className="mr-1">
                    {label}
                  </Badge>
                )),
              },
            ]}
          />
        </Section>

        <Section title="Relationships">
          <RelationList
            entity={selectedEntity}
            relations={selectRelations({ entity: selectedEntity })}
          />
        </Section>

        <Section title="Recent activity">
          <ActivityFeed
            entries={selectActivity({ entity: selectedEntity })}
            people={mockPeople}
          />
        </Section>
      </div>

      <Section title="Entity picker" description="Shared identity selection">
        <div className="max-w-md">
          <EntitySelect
            aria-label="Entity"
            options={mockWorkItems.slice(0, 4).map((item) => ({
              value: item.id,
              title: `${item.key} ${item.title}`,
              description: item.summary,
            }))}
            onValueChange={() => undefined}
          />
        </div>
      </Section>
    </div>
  ),
};

export const Empty: Story = {
  render: () => (
    <div className="mx-auto grid max-w-5xl gap-6 p-6 md:grid-cols-2">
      <AppEmptyState
        icon={<BellRingIcon />}
        title="You’re all caught up"
        description="New attention signals will appear here when work needs a response."
      />
      <AppEmptyState
        icon={<Link2Icon />}
        title="No relationships"
        description="Connect this entity to work or documents to build its context."
        action={
          <Button variant="outline">
            <PlusIcon />
            Add relationship
          </Button>
        }
      />
    </div>
  ),
};

export const Dense: Story = {
  render: () => (
    <div className="mx-auto max-w-3xl p-6">
      <Section
        title="Dense property set"
        description="Compact inspector-oriented presentation"
      >
        <PropertyList
          compact
          items={Array.from({ length: 18 }, (_, index) => ({
            label: `Property ${index + 1}`,
            value:
              index % 3 === 0 ? (
                <StatusIndicator
                  status={index % 2 === 0 ? "active" : "blocked"}
                />
              ) : (
                `Operational value ${index + 1}`
              ),
          }))}
        />
      </Section>
    </div>
  ),
};
