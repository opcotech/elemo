import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";

import { withRouter } from "../../.storybook/with-router";

import { WorkSurface } from "@/components/work";
import { WorkBoard as Board } from "@/components/work/work-board";
import { CompactWorkList } from "@/components/work/work-list";
import { WorkTable } from "@/components/work/work-table";
import { WorkTimeline } from "@/components/work/work-timeline";
import { mockWorkItems } from "@/lib/mock-data";
import type { WorkItem } from "@/lib/mock-data";
import type { WorkRouteSearch } from "@/lib/work-route-search";

const defaultSearch: WorkRouteSearch = {
  display: "comfortable",
  group: "status",
  layout: "board",
  sort: "rank:asc",
};

function StatefulWorkSurface({
  initialSearch = defaultSearch,
}: {
  initialSearch?: WorkRouteSearch;
}) {
  const [search, setSearch] = useState(initialSearch);
  return (
    <div className="h-205 overflow-hidden">
      <WorkSurface
        title="Product delivery"
        description="One query projected as board, list, table, or timeline."
        context={{ namespace: "Product", project: "Elemo web" }}
        scope={{
          type: "project",
          namespaceId: "namespace-product",
          projectId: "project-web",
        }}
        search={search}
        onSearchChange={(patch) =>
          setSearch((previous) => ({ ...previous, ...patch }))
        }
      />
    </div>
  );
}

const denseItems: WorkItem[] = Array.from({ length: 30 }, (_, index) => {
  const source = mockWorkItems[index % mockWorkItems.length];
  return {
    ...source,
    id: `${source.id}-dense-${index}`,
    key: `DENSE-${String(index + 1).padStart(2, "0")}`,
    rank: index + 1,
    status: (["backlog", "planned", "in-progress", "blocked", "done"] as const)[
      index % 5
    ],
    title: `${source.title} · scenario ${index + 1}`,
  };
});

const meta: Meta = {
  title: "Elemo/Work Surface",
  decorators: [withRouter],
  parameters: {
    a11y: { test: "error" },
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "Operational work compositions demonstrating shared query state, projections, density, responsive behavior, inspector context, and theme treatment.",
      },
    },
  },
  tags: ["autodocs", "verification"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Populated: Story = {
  render: () => <StatefulWorkSurface />,
};

export const Empty: Story = {
  render: () => (
    <StatefulWorkSurface
      initialSearch={{
        ...defaultSearch,
        filter: "query-that-matches-no-work",
        layout: "list",
      }}
    />
  ),
};

export const Dense: Story = {
  render: () => (
    <div className="bg-background h-190 overflow-hidden p-4">
      <h1 className="sr-only">Dense work board</h1>
      <Board
        items={denseItems}
        group="status"
        compact
        onSelect={() => undefined}
      />
    </div>
  ),
};

export const ListProjection: Story = {
  render: () => (
    <div className="bg-background mx-auto max-w-3xl p-4">
      <h1 className="sr-only">Work list projection</h1>
      <CompactWorkList
        items={mockWorkItems.slice(0, 8)}
        onSelect={() => undefined}
      />
    </div>
  ),
};

export const TableProjection: Story = {
  render: () => (
    <div className="bg-background h-150 overflow-hidden p-4">
      <h1 className="sr-only">Work table projection</h1>
      <WorkTable
        items={mockWorkItems.slice(0, 12)}
        compact={false}
        onSelect={() => undefined}
      />
    </div>
  ),
};

export const TimelineProjection: Story = {
  render: () => (
    <div className="bg-background h-150 overflow-auto p-4">
      <h1 className="sr-only">Work timeline projection</h1>
      <WorkTimeline
        items={mockWorkItems.slice(0, 10)}
        scope={{
          type: "project",
          namespaceId: "namespace-product",
          projectId: "project-web",
        }}
        compact={false}
        onSelect={() => undefined}
      />
    </div>
  ),
};

export const Mobile: Story = {
  parameters: {
    viewport: { defaultViewport: "mobile1" },
  },
  render: () => (
    <div className="bg-background mx-auto h-190 w-97.5 max-w-full overflow-hidden border-x">
      <StatefulWorkSurface
        initialSearch={{ ...defaultSearch, layout: "list" }}
      />
    </div>
  ),
};

export const InspectorOpen: Story = {
  render: () => (
    <StatefulWorkSurface
      initialSearch={{
        ...defaultSearch,
        layout: "table",
        selected: `work:${mockWorkItems[0].id}`,
      }}
    />
  ),
};

export const Dark: Story = {
  globals: { theme: "dark" },
  render: () => (
    <StatefulWorkSurface
      initialSearch={{
        ...defaultSearch,
        display: "compact",
        layout: "timeline",
      }}
    />
  ),
};
