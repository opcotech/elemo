import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";

import { withRouter } from "../../.storybook/with-router";

import { ResponsiveInspectorShell } from "@/components/layout/responsive-inspector-shell";
import { PropertyList } from "@/components/shared/property-list";
import { StatusIndicator } from "@/components/shared/status-indicator";
import { Button } from "@/components/ui/button";
import { WorkInspector } from "@/components/work/work-inspector";
import { mockWorkItems } from "@/lib/mock-data";

const sampleItem = mockWorkItems[0];

function InspectorDemo({ initiallyOpen = true }: { initiallyOpen?: boolean }) {
  const [open, setOpen] = useState(initiallyOpen);

  return (
    <div className="bg-background h-150 w-full overflow-hidden border">
      <ResponsiveInspectorShell
        open={open}
        onOpenChange={setOpen}
        inspectorTitle={sampleItem.key}
        inspectorDescription={sampleItem.title}
        inspector={open ? <WorkInspector item={sampleItem} /> : undefined}
      >
        <div className="space-y-4 p-4">
          <h1 className="text-lg font-semibold">Work projection</h1>
          <p className="text-muted-foreground text-sm">
            All viewports open a right-side Sheet over the projection, with
            labelled close controls and focus management.
          </p>
          <PropertyList
            items={[
              {
                label: "Selected",
                value: open ? sampleItem.title : "None",
              },
              {
                label: "Status",
                value: <StatusIndicator status={sampleItem.status} />,
              },
            ]}
          />
          <Button type="button" onClick={() => setOpen((value) => !value)}>
            {open ? "Close inspector" : "Open inspector"}
          </Button>
        </div>
      </ResponsiveInspectorShell>
    </div>
  );
}

const meta: Meta = {
  title: "Elemo/Responsive Inspector",
  decorators: [withRouter],
  parameters: {
    a11y: { test: "error" },
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "Responsive inspector shell: right-side Sheet (550–750px on desktop, full-width on smaller viewports) with labelled close controls and focus management.",
      },
    },
  },
  tags: ["autodocs", "verification"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Open: Story = {
  render: () => <InspectorDemo />,
};

export const Closed: Story = {
  render: () => <InspectorDemo initiallyOpen={false} />,
};

export const Mobile: Story = {
  parameters: {
    viewport: { defaultViewport: "mobile1" },
  },
  render: () => (
    <div className="mx-auto w-97.5 max-w-full">
      <InspectorDemo />
    </div>
  ),
};
