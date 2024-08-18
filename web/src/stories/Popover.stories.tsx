import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  Calendar,
  Download,
  Edit,
  HelpCircle,
  Info,
  Mail,
  MapPin,
  MoreHorizontal,
  Phone,
  Plus,
  Settings,
  Share,
  Trash2,
  User,
} from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";

const meta: Meta<typeof Popover> = {
  title: "UI/Popover",
  component: Popover,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Displays rich content in a portal, triggered by a button. Built on top of Radix UI Popover primitive.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    open: {
      control: "boolean",
      description: "Controls the open state of the popover",
    },
    onOpenChange: {
      action: "onOpenChange",
      description: "Callback fired when the open state changes",
    },
    modal: {
      control: "boolean",
      description: "Whether the popover is modal",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic popover
export const Default: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline">Open popover</Button>
      </PopoverTrigger>
      <PopoverContent className="w-80">
        <div className="grid gap-4">
          <div className="space-y-2">
            <h4 className="leading-none font-medium">Dimensions</h4>
            <p className="text-muted-foreground text-sm">
              Set the dimensions for the layer.
            </p>
          </div>
          <div className="grid gap-2">
            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="width">Width</Label>
              <Input
                id="width"
                defaultValue="100%"
                className="col-span-2 h-8"
              />
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="maxWidth">Max. width</Label>
              <Input
                id="maxWidth"
                defaultValue="300px"
                className="col-span-2 h-8"
              />
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="height">Height</Label>
              <Input
                id="height"
                defaultValue="25px"
                className="col-span-2 h-8"
              />
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="maxHeight">Max. height</Label>
              <Input
                id="maxHeight"
                defaultValue="none"
                className="col-span-2 h-8"
              />
            </div>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  ),
};

// Simple popover
export const Simple: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline">
          <Info className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent>
        <p className="text-sm">
          This is a simple popover with basic information.
        </p>
      </PopoverContent>
    </Popover>
  ),
};

// User profile popover
export const UserProfile: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" className="h-8 w-8 rounded-full">
          <User className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80">
        <div className="flex gap-4">
          <div className="bg-muted flex h-12 w-12 items-center justify-center rounded-full">
            <User className="h-6 w-6" />
          </div>
          <div className="grid gap-1">
            <h4 className="font-semibold">Sarah Johnson</h4>
            <p className="text-muted-foreground text-sm">
              Product Designer at Acme Inc.
            </p>
            <div className="text-muted-foreground flex items-center gap-2 text-xs">
              <Mail className="h-3 w-3" />
              sarah@acme.com
            </div>
            <div className="text-muted-foreground flex items-center gap-2 text-xs">
              <MapPin className="h-3 w-3" />
              San Francisco, CA
            </div>
          </div>
        </div>
        <Separator className="my-4" />
        <div className="flex gap-2">
          <Button size="sm" className="flex-1">
            <Mail className="h-3 w-3" />
            Message
          </Button>
          <Button size="sm" variant="outline" className="flex-1">
            <Phone className="h-3 w-3" />
            Call
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  ),
};

// Form popover
export const FormPopover: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button>
          <Plus className="h-4 w-4" />
          Add Note
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80">
        <div className="grid gap-4">
          <div className="space-y-2">
            <h4 className="leading-none font-medium">Add Note</h4>
            <p className="text-muted-foreground text-sm">
              Create a new note with title and description.
            </p>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="title">Title</Label>
            <Input id="title" placeholder="Enter note title" />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              placeholder="Enter note description"
              rows={3}
            />
          </div>
          <div className="flex gap-2">
            <Button size="sm" className="flex-1">
              Save Note
            </Button>
            <Button size="sm" variant="outline" className="flex-1">
              Cancel
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  ),
};

// Actions menu popover
export const ActionsMenu: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="sm">
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-56">
        <div className="grid gap-1">
          <Button variant="ghost" className="h-8 justify-start">
            <Edit className="h-3 w-3" />
            Edit
          </Button>
          <Button variant="ghost" className="h-8 justify-start">
            <Share className="h-3 w-3" />
            Share
          </Button>
          <Button variant="ghost" className="h-8 justify-start">
            <Download className="h-3 w-3" />
            Download
          </Button>
          <Separator className="my-1" />
          <Button variant="ghost" className="h-8 justify-start text-red-600">
            <Trash2 className="h-3 w-3" />
            Delete
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  ),
};

// Help popover
export const Help: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <Label>Email Address</Label>
      <Popover>
        <PopoverTrigger asChild>
          <Button variant="ghost" size="sm" className="h-4 w-4 p-0">
            <HelpCircle className="h-3 w-3" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-72">
          <div className="space-y-2">
            <h4 className="font-medium">Email Address Help</h4>
            <p className="text-muted-foreground text-sm">
              Your email address is used for account notifications and password
              recovery. We'll never share your email with third parties.
            </p>
            <ul className="text-muted-foreground space-y-1 text-sm">
              <li>• Must be a valid email format</li>
              <li>• Cannot be changed once set</li>
              <li>• Used for important account updates</li>
            </ul>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  ),
};

// Calendar popover
export const CalendarPopover: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline">
          <Calendar className="h-4 w-4" />
          Pick a date
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0">
        <div className="p-3">
          <div className="grid gap-2">
            <div className="text-sm font-medium">Select Date</div>
            <div className="grid grid-cols-7 gap-1 text-center text-xs">
              <div className="p-2 font-medium">Sun</div>
              <div className="p-2 font-medium">Mon</div>
              <div className="p-2 font-medium">Tue</div>
              <div className="p-2 font-medium">Wed</div>
              <div className="p-2 font-medium">Thu</div>
              <div className="p-2 font-medium">Fri</div>
              <div className="p-2 font-medium">Sat</div>
              {Array.from({ length: 35 }, (_, i) => (
                <Button
                  key={i}
                  variant="ghost"
                  size="sm"
                  className="h-8 w-8 p-0 text-xs"
                >
                  {i % 7 === 0
                    ? Math.floor(i / 7) + 1
                    : (i % 7) + Math.floor(i / 7) * 7 + 1}
                </Button>
              ))}
            </div>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  ),
};

// Settings popover
export const SettingsPopover: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm">
          <Settings className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-64">
        <div className="grid gap-4">
          <div className="space-y-2">
            <h4 className="leading-none font-medium">Quick Settings</h4>
            <p className="text-muted-foreground text-sm">
              Adjust your preferences.
            </p>
          </div>
          <div className="grid gap-3">
            <div className="flex items-center justify-between">
              <Label className="text-sm">Notifications</Label>
              <Button variant="outline" size="sm">
                On
              </Button>
            </div>
            <div className="flex items-center justify-between">
              <Label className="text-sm">Auto-save</Label>
              <Button variant="outline" size="sm">
                Off
              </Button>
            </div>
            <div className="flex items-center justify-between">
              <Label className="text-sm">Theme</Label>
              <Button variant="outline" size="sm">
                Dark
              </Button>
            </div>
          </div>
          <Separator />
          <Button size="sm" variant="outline" className="w-full">
            Advanced Settings
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  ),
};

// Controlled popover
export const Controlled: Story = {
  render: () => {
    const [open, setOpen] = useState(false);

    return (
      <div className="space-y-2">
        <div className="text-muted-foreground text-sm">
          Popover is {open ? "open" : "closed"}
        </div>
        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger asChild>
            <Button variant="outline">Controlled Popover</Button>
          </PopoverTrigger>
          <PopoverContent>
            <div className="space-y-2">
              <p className="text-sm">
                This popover is controlled by React state.
              </p>
              <Button size="sm" onClick={() => setOpen(false)}>
                Close
              </Button>
            </div>
          </PopoverContent>
        </Popover>
      </div>
    );
  },
};

// Multiple popovers
export const MultiplePopovers: Story = {
  render: () => (
    <div className="flex gap-4">
      <Popover>
        <PopoverTrigger asChild>
          <Button>Popover 1</Button>
        </PopoverTrigger>
        <PopoverContent>
          <p className="text-sm">This is the first popover.</p>
        </PopoverContent>
      </Popover>

      <Popover>
        <PopoverTrigger asChild>
          <Button variant="outline">Popover 2</Button>
        </PopoverTrigger>
        <PopoverContent>
          <p className="text-sm">This is the second popover.</p>
        </PopoverContent>
      </Popover>

      <Popover>
        <PopoverTrigger asChild>
          <Button variant="secondary">Popover 3</Button>
        </PopoverTrigger>
        <PopoverContent>
          <p className="text-sm">This is the third popover.</p>
        </PopoverContent>
      </Popover>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Multiple independent popover instances.",
      },
    },
  },
};

// Different positions
export const Positioning: Story = {
  render: () => (
    <div className="grid grid-cols-3 gap-4 p-8">
      <Popover>
        <PopoverTrigger asChild>
          <Button variant="outline">Top</Button>
        </PopoverTrigger>
        <PopoverContent side="top">
          <p className="text-sm">Popover positioned on top</p>
        </PopoverContent>
      </Popover>

      <Popover>
        <PopoverTrigger asChild>
          <Button variant="outline">Right</Button>
        </PopoverTrigger>
        <PopoverContent side="right">
          <p className="text-sm">Popover positioned on right</p>
        </PopoverContent>
      </Popover>

      <Popover>
        <PopoverTrigger asChild>
          <Button variant="outline">Bottom</Button>
        </PopoverTrigger>
        <PopoverContent side="bottom">
          <p className="text-sm">Popover positioned on bottom</p>
        </PopoverContent>
      </Popover>

      <div></div>

      <Popover>
        <PopoverTrigger asChild>
          <Button variant="outline">Left</Button>
        </PopoverTrigger>
        <PopoverContent side="left">
          <p className="text-sm">Popover positioned on left</p>
        </PopoverContent>
      </Popover>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Popovers positioned on different sides of the trigger.",
      },
    },
  },
};
