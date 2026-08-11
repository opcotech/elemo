import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  ArrowRight,
  Download,
  Edit,
  Heart,
  Plus,
  Save,
  Search,
  Settings,
  Star,
  Trash2,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

const meta: Meta<typeof Button> = {
  title: "UI/Button",
  component: Button,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A versatile button component with multiple variants and sizes. Built on Base UI with Nova styling and motion-friendly press feedback.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    variant: {
      control: "select",
      options: [
        "default",
        "outline",
        "secondary",
        "ghost",
        "destructive",
        "destructive-ghost",
        "link",
        "success",
        "warning",
      ],
      description: "The visual style variant of the button",
    },
    size: {
      control: "select",
      options: [
        "xs",
        "sm",
        "default",
        "lg",
        "icon",
        "icon-xs",
        "icon-sm",
        "icon-lg",
      ],
      description: "The size of the button",
    },
    disabled: {
      control: "boolean",
      description: "Whether the button is disabled",
    },
  },
  args: {},
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    children: "Default Button",
  },
};

export const Destructive: Story = {
  args: {
    variant: "destructive",
    children: "Delete Account",
  },
};

export const DestructiveGhost: Story = {
  args: {
    variant: "destructive-ghost",
    children: "Remove",
  },
};

export const Outline: Story = {
  args: {
    variant: "outline",
    children: "Outline Button",
  },
};

export const Secondary: Story = {
  args: {
    variant: "secondary",
    children: "Secondary Button",
  },
};

export const Ghost: Story = {
  args: {
    variant: "ghost",
    children: "Ghost Button",
  },
};

export const Link: Story = {
  args: {
    variant: "link",
    children: "Link Button",
  },
};

export const Success: Story = {
  args: {
    variant: "success",
    children: "Confirm",
  },
};

export const Warning: Story = {
  args: {
    variant: "warning",
    children: "Proceed with caution",
  },
};

export const ExtraSmall: Story = {
  args: {
    size: "xs",
    children: "Extra Small",
  },
};

export const Small: Story = {
  args: {
    size: "sm",
    children: "Small Button",
  },
};

export const Large: Story = {
  args: {
    size: "lg",
    children: "Large Button",
  },
};

export const Icon: Story = {
  args: {
    size: "icon",
    children: <Settings className="h-4 w-4" />,
  },
};

export const WithIconLeft: Story = {
  args: {
    children: (
      <>
        <Download className="h-4 w-4" />
        Download
      </>
    ),
  },
};

export const WithIconRight: Story = {
  args: {
    children: (
      <>
        Continue
        <ArrowRight className="h-4 w-4" />
      </>
    ),
  },
};

export const IconOnly: Story = {
  args: {
    size: "icon",
    variant: "outline",
    children: <Heart className="h-4 w-4" />,
    "aria-label": "Like",
  },
};

export const Disabled: Story = {
  args: {
    disabled: true,
    children: "Disabled Button",
  },
};

export const Loading: Story = {
  args: {
    disabled: true,
    children: (
      <>
        <Spinner size="xs" className="mr-2" />
        Loading...
      </>
    ),
  },
};

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-wrap gap-3">
      <Button variant="default">Default</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="destructive">Destructive</Button>
      <Button variant="destructive-ghost">Destructive Ghost</Button>
      <Button variant="link">Link</Button>
      <Button variant="success">Success</Button>
      <Button variant="warning">Warning</Button>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "All available button variants displayed together.",
      },
    },
  },
};

export const AllSizes: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      <Button size="xs">Extra Small</Button>
      <Button size="sm">Small</Button>
      <Button size="default">Default</Button>
      <Button size="lg">Large</Button>
      <Button size="icon-xs">
        <Star className="h-3 w-3" />
      </Button>
      <Button size="icon-sm">
        <Star className="h-3.5 w-3.5" />
      </Button>
      <Button size="icon">
        <Star className="h-4 w-4" />
      </Button>
      <Button size="icon-lg">
        <Star className="h-4 w-4" />
      </Button>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "All available button sizes displayed together.",
      },
    },
  },
};

export const ActionButtons: Story = {
  render: () => (
    <div className="flex flex-wrap gap-3">
      <Button>
        <Plus className="h-4 w-4" />
        Add New
      </Button>
      <Button variant="outline">
        <Edit className="h-4 w-4" />
        Edit
      </Button>
      <Button variant="secondary">
        <Save className="h-4 w-4" />
        Save
      </Button>
      <Button variant="destructive">
        <Trash2 className="h-4 w-4" />
        Delete
      </Button>
      <Button variant="ghost" size="icon">
        <Search className="h-4 w-4" />
      </Button>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Common action buttons with appropriate icons and variants.",
      },
    },
  },
};

export const AsLink: Story = {
  render: () => (
    <Button render={<a href="#" role="button" />}>Link styled as button</Button>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Using the `render` prop to apply button styling to a different element.",
      },
    },
  },
};
