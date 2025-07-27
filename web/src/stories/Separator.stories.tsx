import type { Meta, StoryObj } from "@storybook/react-vite";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";

const meta: Meta<typeof Separator> = {
  title: "UI/Separator",
  component: Separator,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Visually or semantically separates content. A simple line that can be horizontal or vertical to divide content sections.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    orientation: {
      control: "select",
      options: ["horizontal", "vertical"],
      description: "The orientation of the separator",
    },
    decorative: {
      control: "boolean",
      description:
        "Whether the separator is decorative (true) or semantic (false)",
    },
    className: {
      control: "text",
      description: "Additional CSS classes to apply to the separator",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic horizontal separator
export const Default: Story = {
  render: () => (
    <div className="w-64 space-y-4">
      <div className="space-y-1">
        <h4 className="text-sm leading-none font-medium">Radix Primitives</h4>
        <p className="text-muted-foreground text-sm">
          An open-source UI component library.
        </p>
      </div>
      <Separator />
      <div className="flex h-5 items-center space-x-4 text-sm">
        <div>Blog</div>
        <Separator orientation="vertical" />
        <div>Docs</div>
        <Separator orientation="vertical" />
        <div>Source</div>
      </div>
    </div>
  ),
};

// Horizontal separator
export const Horizontal: Story = {
  render: () => (
    <div className="w-80 space-y-4">
      <div>
        <h3 className="text-lg font-medium">Section 1</h3>
        <p className="text-muted-foreground text-sm">
          This is the first section of content.
        </p>
      </div>
      <Separator />
      <div>
        <h3 className="text-lg font-medium">Section 2</h3>
        <p className="text-muted-foreground text-sm">
          This is the second section of content.
        </p>
      </div>
      <Separator />
      <div>
        <h3 className="text-lg font-medium">Section 3</h3>
        <p className="text-muted-foreground text-sm">
          This is the third section of content.
        </p>
      </div>
    </div>
  ),
};

// Vertical separator
export const Vertical: Story = {
  render: () => (
    <div className="flex h-20 items-center space-x-4">
      <div className="text-center">
        <div className="text-lg font-semibold">Home</div>
        <div className="text-muted-foreground text-xs">Main page</div>
      </div>
      <Separator orientation="vertical" />
      <div className="text-center">
        <div className="text-lg font-semibold">About</div>
        <div className="text-muted-foreground text-xs">Learn more</div>
      </div>
      <Separator orientation="vertical" />
      <div className="text-center">
        <div className="text-lg font-semibold">Contact</div>
        <div className="text-muted-foreground text-xs">Get in touch</div>
      </div>
    </div>
  ),
};

// Navigation with separators
export const Navigation: Story = {
  render: () => (
    <nav className="flex items-center space-x-4 text-sm">
      <Button variant="ghost" size="sm">
        Home
      </Button>
      <Separator orientation="vertical" className="h-4" />
      <Button variant="ghost" size="sm">
        Products
      </Button>
      <Separator orientation="vertical" className="h-4" />
      <Button variant="ghost" size="sm">
        About
      </Button>
      <Separator orientation="vertical" className="h-4" />
      <Button variant="ghost" size="sm">
        Contact
      </Button>
    </nav>
  ),
};

// Form sections
export const FormSections: Story = {
  render: () => (
    <div className="w-80 space-y-6">
      <div className="space-y-4">
        <h3 className="text-lg font-medium">Personal Information</h3>
        <div className="grid gap-4">
          <div className="grid gap-2">
            <Label htmlFor="name">Full Name</Label>
            <Input id="name" placeholder="John Doe" />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="email">Email</Label>
            <Input id="email" type="email" placeholder="john@example.com" />
          </div>
        </div>
      </div>

      <Separator />

      <div className="space-y-4">
        <h3 className="text-lg font-medium">Account Settings</h3>
        <div className="grid gap-4">
          <div className="grid gap-2">
            <Label htmlFor="username">Username</Label>
            <Input id="username" placeholder="johndoe" />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="password">Password</Label>
            <Input id="password" type="password" placeholder="••••••••" />
          </div>
        </div>
      </div>

      <Separator />

      <div className="space-y-4">
        <h3 className="text-lg font-medium">Preferences</h3>
        <div className="flex items-center justify-between">
          <Label>Email notifications</Label>
          <Button variant="outline" size="sm">
            Enabled
          </Button>
        </div>
      </div>
    </div>
  ),
};

// Breadcrumb separator
export const Breadcrumb: Story = {
  render: () => (
    <nav className="text-muted-foreground flex items-center space-x-2 text-sm">
      <Button variant="link" className="text-muted-foreground h-auto p-0">
        Home
      </Button>
      <Separator orientation="vertical" className="h-4" />
      <Button variant="link" className="text-muted-foreground h-auto p-0">
        Products
      </Button>
      <Separator orientation="vertical" className="h-4" />
      <Button variant="link" className="text-muted-foreground h-auto p-0">
        Electronics
      </Button>
      <Separator orientation="vertical" className="h-4" />
      <span className="text-foreground">Smartphones</span>
    </nav>
  ),
};

// Card sections
export const CardSections: Story = {
  render: () => (
    <div className="bg-card w-80 rounded-lg border p-6">
      <div className="flex items-center space-x-4">
        <Avatar>
          <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
          <AvatarFallback>CN</AvatarFallback>
        </Avatar>
        <div>
          <h4 className="text-sm font-semibold">John Doe</h4>
          <p className="text-muted-foreground text-sm">Software Engineer</p>
        </div>
      </div>

      <Separator className="my-4" />

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-sm">Status</span>
          <Badge variant="secondary">Active</Badge>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-sm">Team</span>
          <span className="text-muted-foreground text-sm">Frontend</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-sm">Location</span>
          <span className="text-muted-foreground text-sm">San Francisco</span>
        </div>
      </div>

      <Separator className="my-4" />

      <div className="flex space-x-2">
        <Button size="sm" className="flex-1">
          Message
        </Button>
        <Button size="sm" variant="outline" className="flex-1">
          Call
        </Button>
      </div>
    </div>
  ),
};

// Stats dashboard
export const StatsDashboard: Story = {
  render: () => (
    <div className="w-96 space-y-6">
      <h2 className="text-xl font-semibold">Analytics Dashboard</h2>

      <div className="grid grid-cols-3 gap-4">
        <div className="text-center">
          <div className="text-2xl font-bold">1,234</div>
          <div className="text-muted-foreground text-sm">Users</div>
        </div>
        <Separator orientation="vertical" className="justify-self-center" />
        <div className="text-center">
          <div className="text-2xl font-bold">5,678</div>
          <div className="text-muted-foreground text-sm">Orders</div>
        </div>
      </div>

      <Separator />

      <div className="grid grid-cols-3 gap-4">
        <div className="text-center">
          <div className="text-2xl font-bold">$12.3K</div>
          <div className="text-muted-foreground text-sm">Revenue</div>
        </div>
        <Separator orientation="vertical" className="justify-self-center" />
        <div className="text-center">
          <div className="text-2xl font-bold">89%</div>
          <div className="text-muted-foreground text-sm">Conversion</div>
        </div>
      </div>

      <Separator />

      <div className="text-center">
        <div className="text-lg font-medium">Monthly Growth</div>
        <div className="rounded bg-green-600 px-2 text-3xl font-bold text-white">
          +15.3%
        </div>
        <div className="text-muted-foreground text-sm">
          Compared to last month
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Separators used in a dashboard layout to organize statistical information.",
      },
    },
  },
};

// Custom styling
export const CustomStyling: Story = {
  render: () => (
    <div className="w-80 space-y-6">
      <div>
        <h3 className="text-lg font-medium">Default Separator</h3>
        <p className="text-muted-foreground text-sm">Standard appearance</p>
      </div>
      <Separator />

      <div>
        <h3 className="text-lg font-medium">Thick Separator</h3>
        <p className="text-muted-foreground text-sm">With increased height</p>
      </div>
      <Separator className="bg-primary h-1" />

      <div>
        <h3 className="text-lg font-medium">Dashed Separator</h3>
        <p className="text-muted-foreground text-sm">
          With custom border style
        </p>
      </div>
      <Separator className="border-muted-foreground border-t-2 border-dashed" />

      <div>
        <h3 className="text-lg font-medium">Colored Separator</h3>
        <p className="text-muted-foreground text-sm">With custom color</p>
      </div>
      <Separator className="h-0.5 bg-gradient-to-r from-blue-500 to-purple-500" />

      <div>
        <h3 className="text-lg font-medium">Dotted Separator</h3>
        <p className="text-muted-foreground text-sm">With dotted style</p>
      </div>
      <Separator className="border-muted-foreground border-t-2 border-dotted" />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Separators with various custom styling options including thickness, color, and border styles.",
      },
    },
  },
};

// Complex layout
export const ComplexLayout: Story = {
  render: () => (
    <div className="w-96 space-y-4">
      {/* Header */}
      <div className="text-center">
        <h2 className="text-xl font-bold">User Profile</h2>
        <p className="text-muted-foreground text-sm">
          Manage your account settings
        </p>
      </div>

      <Separator />

      {/* Profile section */}
      <div className="flex items-center space-x-4">
        <Avatar className="h-16 w-16">
          <AvatarImage src="https://github.com/shadcn.png" alt="Profile" />
          <AvatarFallback>JD</AvatarFallback>
        </Avatar>
        <div className="flex-1">
          <h3 className="font-medium">John Doe</h3>
          <p className="text-muted-foreground text-sm">john.doe@example.com</p>
          <div className="mt-1 flex items-center space-x-2">
            <Badge variant="secondary">Pro</Badge>
            <Separator orientation="vertical" className="h-4" />
            <span className="text-muted-foreground text-xs">
              Member since 2023
            </span>
          </div>
        </div>
      </div>

      <Separator />

      {/* Quick stats */}
      <div className="grid grid-cols-3 gap-4 text-center">
        <div>
          <div className="text-lg font-semibold">142</div>
          <div className="text-muted-foreground text-xs">Posts</div>
        </div>
        <Separator orientation="vertical" className="justify-self-center" />
        <div>
          <div className="text-lg font-semibold">1.2K</div>
          <div className="text-muted-foreground text-xs">Followers</div>
        </div>
        <Separator orientation="vertical" className="justify-self-center" />
        <div>
          <div className="text-lg font-semibold">342</div>
          <div className="text-muted-foreground text-xs">Following</div>
        </div>
      </div>

      <Separator />

      {/* Actions */}
      <div className="flex space-x-2">
        <Button className="flex-1">Edit Profile</Button>
        <Button variant="outline" className="flex-1">
          Settings
        </Button>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "A complex layout using multiple separators to organize different sections of content.",
      },
    },
  },
};
