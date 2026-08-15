import type { Meta, StoryObj } from "@storybook/react-vite";

import { withRouter } from "../../.storybook/with-router";

import { InternalLink } from "@/components/ui/internal-link";
import { internalPath } from "@/lib/internal-url";

const meta: Meta<typeof InternalLink> = {
  title: "UI/InternalLink",
  component: InternalLink,
  decorators: [withRouter],
  parameters: {
    a11y: { test: "error" },
    layout: "centered",
    docs: {
      description: {
        component:
          "Typed TanStack Router link for in-app navigation. Prefer this over raw anchors for internal destinations.",
      },
    },
  },
  tags: ["autodocs", "verification"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <InternalLink to={internalPath("/namespaces/demo/work")}>
      Open namespace work
    </InternalLink>
  ),
};

export const InlineInCopy: Story = {
  render: () => (
    <p className="text-sm">
      Continue in{" "}
      <InternalLink to={internalPath("/settings")} className="font-medium">
        Settings
      </InternalLink>{" "}
      to manage your account.
    </p>
  ),
};
