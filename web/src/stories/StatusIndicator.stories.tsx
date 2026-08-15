import type { Meta, StoryObj } from "@storybook/react-vite";

import { StatusIndicator } from "@/components/ui/status-indicator";

const meta: Meta<typeof StatusIndicator> = {
  title: "Elemo/Status Indicator",
  component: StatusIndicator,
  parameters: {
    a11y: { test: "error" },
    layout: "centered",
    docs: {
      description: {
        component:
          "Accessible status as a decorative tone dot plus visible text for work, projects, and priorities.",
      },
    },
  },
  tags: ["autodocs", "verification"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Gallery: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-2">
      <StatusIndicator status="open" />
      <StatusIndicator status="backlog" />
      <StatusIndicator status="in review" />
      <StatusIndicator status="in progress" />
      <StatusIndicator status="review" />
      <StatusIndicator status="done" />
      <StatusIndicator status="blocked" />
      <StatusIndicator status="active" />
      <StatusIndicator status="closed" />
      <StatusIndicator status="high" />
      <StatusIndicator status="medium" />
      <StatusIndicator status="low" />
    </div>
  ),
};

export const ExplicitTone: Story = {
  args: {
    status: "custom",
    tone: "info",
  },
};
