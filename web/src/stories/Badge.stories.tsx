import type { Meta, StoryObj } from "@storybook/react-vite";
import { AlertTriangle, Check, Clock, Info, Star, User, X } from "lucide-react";

import { Badge } from "@/components/ui/badge";

const meta: Meta<typeof Badge> = {
  title: "UI/Badge",
  component: Badge,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A small status indicator or label for categories, statuses, and metadata.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    variant: {
      control: "select",
      options: [
        "default",
        "primary",
        "secondary",
        "destructive",
        "outline",
        "ghost",
        "link",
        "success",
        "warning",
        "info",
      ],
      description: "The visual style variant of the badge",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    children: "Badge",
  },
};

export const Secondary: Story = {
  args: {
    variant: "secondary",
    children: "Secondary",
  },
};

export const Destructive: Story = {
  args: {
    variant: "destructive",
    children: "Destructive",
  },
};

export const Outline: Story = {
  args: {
    variant: "outline",
    children: "Outline",
  },
};

export const Ghost: Story = {
  args: {
    variant: "ghost",
    children: "Ghost",
  },
};

export const Link: Story = {
  args: {
    variant: "link",
    children: "Link",
  },
};

export const Success: Story = {
  args: {
    variant: "success",
    children: "Success",
  },
};

export const Warning: Story = {
  args: {
    variant: "warning",
    children: "Warning",
  },
};

export const InfoVariant: Story = {
  args: {
    variant: "info",
    children: "Info",
  },
};

export const WithIconLeft: Story = {
  args: {
    children: (
      <>
        <Check className="h-3 w-3" />
        Completed
      </>
    ),
    variant: "success",
  },
};

export const WithIconRight: Story = {
  args: {
    children: (
      <>
        Pending
        <Clock className="h-3 w-3" />
      </>
    ),
    variant: "warning",
  },
};

export const IconOnly: Story = {
  args: {
    children: <Star className="h-3 w-3" />,
    variant: "default",
  },
};

export const StatusBadges: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      <Badge variant="success">
        <Check className="h-3 w-3" />
        Active
      </Badge>
      <Badge variant="warning">
        <Clock className="h-3 w-3" />
        Pending
      </Badge>
      <Badge variant="destructive">
        <X className="h-3 w-3" />
        Inactive
      </Badge>
      <Badge variant="info">
        <Info className="h-3 w-3" />
        New
      </Badge>
      <Badge variant="outline">
        <AlertTriangle className="h-3 w-3" />
        Draft
      </Badge>
    </div>
  ),
};

export const CategoryBadges: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      <Badge variant="default">Technology</Badge>
      <Badge variant="secondary">Design</Badge>
      <Badge variant="outline">Marketing</Badge>
    </div>
  ),
};

export const UserBadges: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      <Badge variant="default">
        <User className="h-3 w-3" />
        Admin
      </Badge>
      <Badge variant="secondary">
        <Star className="h-3 w-3" />
        Pro User
      </Badge>
      <Badge variant="outline">Member</Badge>
    </div>
  ),
};

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      <Badge variant="default">Default</Badge>
      <Badge variant="primary">Primary</Badge>
      <Badge variant="secondary">Secondary</Badge>
      <Badge variant="destructive">Destructive</Badge>
      <Badge variant="outline">Outline</Badge>
      <Badge variant="ghost">Ghost</Badge>
      <Badge variant="link">Link</Badge>
      <Badge variant="success">Success</Badge>
      <Badge variant="warning">Warning</Badge>
      <Badge variant="info">Info</Badge>
    </div>
  ),
};

export const ContentTypes: Story = {
  render: () => (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-2">
        <Badge>123</Badge>
        <Badge variant="secondary">New</Badge>
        <Badge variant="success">✓ Verified</Badge>
        <Badge variant="warning">2 days left</Badge>
      </div>
      <div className="flex flex-wrap gap-2">
        <Badge variant="success">Online</Badge>
        <Badge variant="destructive">99+</Badge>
      </div>
    </div>
  ),
};

export const AsLink: Story = {
  render: () => (
    <Badge
      variant="outline"
      render={<a href="#" className="hover:no-underline" />}
    >
      Clickable Badge
    </Badge>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Using the `render` prop to render the badge as a clickable link.",
      },
    },
  },
};

export const NotificationBadges: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <div className="relative">
        <button className="bg-secondary rounded-lg p-2">
          <User className="h-5 w-5" />
        </button>
        <Badge className="absolute -top-1 -right-1 h-5 min-w-[1.25rem] px-1">
          3
        </Badge>
      </div>
      <div className="relative">
        <button className="bg-secondary rounded-lg p-2">Messages</button>
        <Badge
          variant="destructive"
          className="absolute -top-2 -right-2 h-5 min-w-[1.25rem] px-1"
        >
          99+
        </Badge>
      </div>
    </div>
  ),
};
