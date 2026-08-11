import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  AlertTriangle,
  Calendar,
  CheckCircle,
  Clock,
  Copy,
  Download,
  Edit,
  HelpCircle,
  Info,
  MapPin,
  Plus,
  Settings,
  Share,
  Star,
  Trash2,
  User,
  XCircle,
} from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

const meta: Meta<typeof Tooltip> = {
  title: "UI/Tooltip",
  component: Tooltip,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A tooltip component that displays helpful information when hovering over or focusing on an element. Built on top of Radix UI with smooth animations and customizable positioning.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    open: {
      control: "boolean",
      description: "Whether the tooltip is open",
    },
    defaultOpen: {
      control: "boolean",
      description: "Whether the tooltip is open by default",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic tooltip
export const Default: Story = {
  render: () => (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger render={<Button variant="outline" />}>
          Hover me
        </TooltipTrigger>
        <TooltipContent>
          <p>This is a helpful tooltip</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  ),
};

// Tooltip with icon button
export const WithIconButton: Story = {
  render: () => (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger render={<Button variant="outline" size="icon" />}>
          <Info className="h-4 w-4" />
        </TooltipTrigger>
        <TooltipContent>
          <p>More information</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  ),
};

// Different positions
export const Positioning: Story = {
  render: () => (
    <TooltipProvider>
      <div className="flex flex-col items-center gap-8">
        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" />}>
            Top
          </TooltipTrigger>
          <TooltipContent side="top">
            <p>Tooltip on top</p>
          </TooltipContent>
        </Tooltip>

        <div className="flex gap-8">
          <Tooltip>
            <TooltipTrigger render={<Button variant="outline" />}>
              Left
            </TooltipTrigger>
            <TooltipContent side="left">
              <p>Tooltip on left</p>
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger render={<Button variant="outline" />}>
              Right
            </TooltipTrigger>
            <TooltipContent side="right">
              <p>Tooltip on right</p>
            </TooltipContent>
          </Tooltip>
        </div>

        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" />}>
            Bottom
          </TooltipTrigger>
          <TooltipContent side="bottom">
            <p>Tooltip on bottom</p>
          </TooltipContent>
        </Tooltip>
      </div>
    </TooltipProvider>
  ),
  parameters: {
    docs: {
      description: {
        story: "Tooltips can be positioned on any side of the trigger element.",
      },
    },
  },
};

// Help tooltips
export const HelpTooltips: Story = {
  render: () => (
    <TooltipProvider>
      <div className="space-y-6 p-4">
        <div className="flex items-center gap-2">
          <span>Username</span>
          <Tooltip>
            <TooltipTrigger
              render={
                <HelpCircle className="text-muted-foreground h-4 w-4 cursor-help" />
              }
            ></TooltipTrigger>
            <TooltipContent>
              <p>
                Choose a unique username that will be visible to other users
              </p>
            </TooltipContent>
          </Tooltip>
        </div>

        <div className="flex items-center gap-2">
          <span>Password strength</span>
          <Tooltip>
            <TooltipTrigger
              render={
                <Info className="text-muted-foreground h-4 w-4 cursor-help" />
              }
            ></TooltipTrigger>
            <TooltipContent className="max-w-xs">
              <div className="space-y-1">
                <p className="font-medium">Password requirements:</p>
                <ul className="space-y-0.5 text-xs">
                  <li>• At least 8 characters</li>
                  <li>• Include uppercase and lowercase letters</li>
                  <li>• Include at least one number</li>
                  <li>• Include at least one special character</li>
                </ul>
              </div>
            </TooltipContent>
          </Tooltip>
        </div>

        <div className="flex items-center gap-2">
          <span>Two-factor authentication</span>
          <Tooltip>
            <TooltipTrigger
              render={
                <AlertTriangle className="h-4 w-4 cursor-help text-amber-500" />
              }
            ></TooltipTrigger>
            <TooltipContent>
              <p>Enable 2FA for enhanced account security</p>
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
    </TooltipProvider>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Tooltips providing helpful information and context for form fields and settings.",
      },
    },
  },
};

// Action tooltips
export const ActionTooltips: Story = {
  render: () => (
    <TooltipProvider>
      <div className="flex gap-2 p-4">
        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" size="icon" />}>
            <Edit className="h-4 w-4" />
          </TooltipTrigger>
          <TooltipContent>
            <p>Edit item</p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" size="icon" />}>
            <Copy className="h-4 w-4" />
          </TooltipTrigger>
          <TooltipContent>
            <p>Copy to clipboard</p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" size="icon" />}>
            <Share className="h-4 w-4" />
          </TooltipTrigger>
          <TooltipContent>
            <p>Share item</p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" size="icon" />}>
            <Download className="h-4 w-4" />
          </TooltipTrigger>
          <TooltipContent>
            <p>Download file</p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger render={<Button variant="destructive" size="icon" />}>
            <Trash2 className="h-4 w-4" />
          </TooltipTrigger>
          <TooltipContent>
            <p>Delete item</p>
          </TooltipContent>
        </Tooltip>
      </div>
    </TooltipProvider>
  ),
  parameters: {
    docs: {
      description: {
        story: "Tooltips for action buttons explaining what each action does.",
      },
    },
  },
};

// Status tooltips
export const StatusTooltips: Story = {
  render: () => (
    <TooltipProvider>
      <div className="flex gap-4 p-4">
        <Tooltip>
          <TooltipTrigger
            render={<div className="flex cursor-help items-center gap-2" />}
          >
            <CheckCircle className="h-5 w-5 text-green-500" />
            <span>Online</span>
          </TooltipTrigger>
          <TooltipContent>
            <p>Service is running normally</p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger
            render={<div className="flex cursor-help items-center gap-2" />}
          >
            <AlertTriangle className="h-5 w-5 text-amber-500" />
            <span>Warning</span>
          </TooltipTrigger>
          <TooltipContent>
            <p>Service is experiencing minor issues</p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger
            render={<div className="flex cursor-help items-center gap-2" />}
          >
            <XCircle className="h-5 w-5 text-red-500" />
            <span>Offline</span>
          </TooltipTrigger>
          <TooltipContent>
            <p>Service is currently unavailable</p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger
            render={<div className="flex cursor-help items-center gap-2" />}
          >
            <Clock className="h-5 w-5 text-blue-500" />
            <span>Pending</span>
          </TooltipTrigger>
          <TooltipContent>
            <p>Service is starting up</p>
          </TooltipContent>
        </Tooltip>
      </div>
    </TooltipProvider>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Tooltips showing detailed status information for various system states.",
      },
    },
  },
};

// Rich content tooltips
export const RichContent: Story = {
  render: () => (
    <TooltipProvider>
      <div className="flex gap-4 p-4">
        <Tooltip>
          <TooltipTrigger render={<Avatar className="cursor-help" />}>
            <AvatarImage src="https://github.com/shadcn.png" />
            <AvatarFallback>CN</AvatarFallback>
          </TooltipTrigger>
          <TooltipContent className="max-w-xs">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Avatar className="h-8 w-8">
                  <AvatarImage src="https://github.com/shadcn.png" />
                  <AvatarFallback>CN</AvatarFallback>
                </Avatar>
                <div>
                  <p className="font-medium">John Doe</p>
                  <p className="text-muted-foreground text-xs">
                    Senior Developer
                  </p>
                </div>
              </div>
              <div className="flex gap-1">
                <Badge variant="secondary" className="text-xs">
                  React
                </Badge>
                <Badge variant="secondary" className="text-xs">
                  TypeScript
                </Badge>
              </div>
            </div>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger
            render={<Button variant="outline" className="cursor-help" />}
          >
            Project Stats
          </TooltipTrigger>
          <TooltipContent className="max-w-sm">
            <div className="space-y-3">
              <h4 className="font-medium">Project Overview</h4>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <p className="text-muted-foreground">Contributors</p>
                  <p className="font-medium">12</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Commits</p>
                  <p className="font-medium">1,247</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Issues</p>
                  <p className="font-medium">23 open</p>
                </div>
                <div>
                  <p className="text-muted-foreground">PRs</p>
                  <p className="font-medium">5 pending</p>
                </div>
              </div>
            </div>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger
            render={<Button variant="outline" className="cursor-help" />}
          >
            <Calendar className="size-4" />
            Event Details
          </TooltipTrigger>
          <TooltipContent className="max-w-xs">
            <div className="space-y-2">
              <h4 className="font-medium">Team Meeting</h4>
              <div className="space-y-1 text-sm">
                <div className="flex items-center gap-2">
                  <Clock className="h-3 w-3" />
                  <span>2:00 PM - 3:00 PM</span>
                </div>
                <div className="flex items-center gap-2">
                  <MapPin className="h-3 w-3" />
                  <span>Conference Room A</span>
                </div>
                <div className="flex items-center gap-2">
                  <User className="h-3 w-3" />
                  <span>8 attendees</span>
                </div>
              </div>
              <p className="text-muted-foreground text-xs">
                Monthly team sync to discuss project progress and upcoming
                milestones.
              </p>
            </div>
          </TooltipContent>
        </Tooltip>
      </div>
    </TooltipProvider>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Tooltips with rich content including avatars, badges, stats, and detailed information.",
      },
    },
  },
};

// Keyboard shortcuts
export const KeyboardShortcuts: Story = {
  render: () => (
    <TooltipProvider>
      <div className="grid grid-cols-2 gap-4 p-4">
        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" />}>
            <Plus className="size-4" />
            New File
          </TooltipTrigger>
          <TooltipContent>
            <div className="flex items-center gap-2">
              <span>Create new file</span>
              <kbd className="bg-muted pointer-events-none inline-flex h-5 items-center gap-1 rounded border px-1.5 font-mono text-[10px] font-medium select-none">
                <span className="text-xs">⌘</span>N
              </kbd>
            </div>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" />}>
            <Copy className="size-4" />
            Search
          </TooltipTrigger>
          <TooltipContent>
            <div className="flex items-center gap-2">
              <span>Search files</span>
              <kbd className="bg-muted pointer-events-none inline-flex h-5 items-center gap-1 rounded border px-1.5 font-mono text-[10px] font-medium select-none">
                <span className="text-xs">⌘</span>K
              </kbd>
            </div>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" />}>
            <Settings className="size-4" />
            Settings
          </TooltipTrigger>
          <TooltipContent>
            <div className="flex items-center gap-2">
              <span>Open settings</span>
              <kbd className="bg-muted pointer-events-none inline-flex h-5 items-center gap-1 rounded border px-1.5 font-mono text-[10px] font-medium select-none">
                <span className="text-xs">⌘</span>,
              </kbd>
            </div>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" />}>
            <Copy className="size-4" />
            Copy
          </TooltipTrigger>
          <TooltipContent>
            <div className="flex items-center gap-2">
              <span>Copy selection</span>
              <kbd className="bg-muted pointer-events-none inline-flex h-5 items-center gap-1 rounded border px-1.5 font-mono text-[10px] font-medium select-none">
                <span className="text-xs">⌘</span>C
              </kbd>
            </div>
          </TooltipContent>
        </Tooltip>
      </div>
    </TooltipProvider>
  ),
  parameters: {
    docs: {
      description: {
        story: "Tooltips showing keyboard shortcuts for various actions.",
      },
    },
  },
};

// Custom delay
export const CustomDelay: Story = {
  render: () => (
    <div className="flex gap-4 p-4">
      <TooltipProvider delay={0}>
        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" />}>
            Instant
          </TooltipTrigger>
          <TooltipContent>
            <p>No delay (0ms)</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>

      <TooltipProvider delay={500}>
        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" />}>
            Normal
          </TooltipTrigger>
          <TooltipContent>
            <p>Normal delay (500ms)</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>

      <TooltipProvider delay={1000}>
        <Tooltip>
          <TooltipTrigger render={<Button variant="outline" />}>
            Slow
          </TooltipTrigger>
          <TooltipContent>
            <p>Slow delay (1000ms)</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Tooltips with different delay durations before appearing.",
      },
    },
  },
};

// Disabled trigger
export const DisabledTrigger: Story = {
  render: () => (
    <TooltipProvider>
      <div className="flex gap-4 p-4">
        <Tooltip>
          <TooltipTrigger render={<span className="inline-block" />}>
            <Button disabled>Disabled Button</Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>This button is currently disabled</p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger
            render={<span className="inline-block cursor-not-allowed" />}
          >
            <Button variant="outline" disabled>
              <Trash2 className="h-4 w-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Delete action is not available</p>
          </TooltipContent>
        </Tooltip>
      </div>
    </TooltipProvider>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Tooltips on disabled elements require wrapping in a span to work properly.",
      },
    },
  },
};

// Complex interactions
export const ComplexInteractions: Story = {
  render: () => (
    <TooltipProvider>
      <div className="space-y-6 p-4">
        <div className="flex items-center gap-2">
          <span>File status:</span>
          <Tooltip>
            <TooltipTrigger
              render={<Badge variant="secondary" className="cursor-help" />}
            >
              Modified
            </TooltipTrigger>
            <TooltipContent>
              <div className="space-y-2">
                <p className="font-medium">File has unsaved changes</p>
                <div className="space-y-1 text-xs">
                  <p>• Last modified: 2 minutes ago</p>
                  <p>• Changes: 15 lines added, 3 deleted</p>
                  <p>• Auto-save: Enabled</p>
                </div>
              </div>
            </TooltipContent>
          </Tooltip>
        </div>

        <div className="flex items-center gap-4">
          <span>User activity:</span>
          <div className="flex -space-x-2">
            {[1, 2, 3].map((i) => (
              <Tooltip key={i}>
                <TooltipTrigger
                  render={
                    <Avatar className="border-background cursor-help border-2" />
                  }
                >
                  <AvatarImage src={`https://github.com/shadcn.png`} />
                  <AvatarFallback>U{i}</AvatarFallback>
                </TooltipTrigger>
                <TooltipContent>
                  <div className="space-y-1">
                    <p className="font-medium">User {i}</p>
                    <p className="text-muted-foreground text-xs">
                      {i === 1
                        ? "Currently editing"
                        : i === 2
                          ? "Viewing"
                          : "Last seen 5m ago"}
                    </p>
                  </div>
                </TooltipContent>
              </Tooltip>
            ))}
          </div>
        </div>

        <div className="flex items-center gap-2">
          <span>Rating:</span>
          <div className="flex">
            {[1, 2, 3, 4, 5].map((star) => (
              <Tooltip key={star}>
                <TooltipTrigger
                  render={
                    <Star
                      className={`h-4 w-4 cursor-pointer ${
                        star <= 4
                          ? "fill-yellow-400 text-yellow-400"
                          : "text-gray-300"
                      }`}
                    />
                  }
                />
                <TooltipContent>
                  <p>
                    {star} star{star !== 1 ? "s" : ""}
                  </p>
                </TooltipContent>
              </Tooltip>
            ))}
          </div>
          <span className="text-muted-foreground ml-2 text-sm">
            4.0 (127 reviews)
          </span>
        </div>
      </div>
    </TooltipProvider>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Complex tooltip interactions with multiple elements and rich information display.",
      },
    },
  },
};
