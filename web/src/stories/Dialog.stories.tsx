import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  Copy,
  Edit,
  Plus,
  SettingsIcon,
  ShareIcon,
  Trash2,
  User,
} from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

const meta: Meta<typeof Dialog> = {
  title: "UI/Dialog",
  component: Dialog,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A window overlaid on either the primary window or another dialog window, rendering the content underneath inert. Built on top of Radix UI Dialog.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    open: {
      control: "boolean",
      description: "Controls the open state of the dialog",
    },
    onOpenChange: {
      action: "onOpenChange",
      description: "Callback fired when the open state changes",
    },
    modal: {
      control: "boolean",
      description: "Whether the dialog is modal",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic dialog
export const Default: Story = {
  render: () => (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline">Edit Profile</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Edit profile</DialogTitle>
          <DialogDescription>
            Make changes to your profile here. Click save when you're done.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="name" className="text-right">
              Name
            </Label>
            <Input
              id="name"
              defaultValue="Pedro Duarte"
              className="col-span-3"
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="username" className="text-right">
              Username
            </Label>
            <Input
              id="username"
              defaultValue="@peduarte"
              className="col-span-3"
            />
          </div>
        </div>
        <DialogFooter>
          <Button type="submit">Save changes</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};

// Simple confirmation dialog
export const Confirmation: Story = {
  render: () => (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="destructive">
          <Trash2 className="h-4 w-4" />
          Delete Item
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Are you sure?</DialogTitle>
          <DialogDescription>
            This action cannot be undone. This will permanently delete the item
            from our servers.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline">Cancel</Button>
          <Button variant="destructive">Delete</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};

// Create new item dialog
export const CreateNew: Story = {
  render: () => (
    <Dialog>
      <DialogTrigger asChild>
        <Button>
          <Plus className="h-4 w-4" />
          Create New Project
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create project</DialogTitle>
          <DialogDescription>
            Create a new project to organize your work.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="project-name">Project name</Label>
            <Input id="project-name" placeholder="Enter project name" />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              placeholder="Describe your project"
              rows={3}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline">Cancel</Button>
          <Button>Create Project</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};

// Share dialog
export const Share: Story = {
  render: () => (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline">
          <ShareIcon className="h-4 w-4" />
          Share
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Share link</DialogTitle>
          <DialogDescription>
            Anyone who has this link will be able to view this.
          </DialogDescription>
        </DialogHeader>
        <div className="flex items-center space-x-2">
          <div className="grid flex-1 gap-2">
            <Label htmlFor="link" className="sr-only">
              Link
            </Label>
            <Input
              id="link"
              defaultValue="https://ui.shadcn.com/docs/installation"
              readOnly
            />
          </div>
          <Button type="button" size="sm" className="px-3">
            <span className="sr-only">Copy</span>
            <Copy className="h-4 w-4" />
          </Button>
        </div>
        <DialogFooter className="sm:justify-start">
          <Button type="button" variant="secondary">
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};

// Settings dialog
export const Settings: Story = {
  render: () => (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline">
          <SettingsIcon className="h-4 w-4" />
          Settings
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Settings</DialogTitle>
          <DialogDescription>
            Manage your application settings and preferences.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-6 py-4">
          <div className="grid gap-3">
            <Label className="text-sm font-medium">Account</Label>
            <div className="grid gap-2">
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="email" className="text-right text-sm">
                  Email
                </Label>
                <Input
                  id="email"
                  defaultValue="user@example.com"
                  className="col-span-3"
                />
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="display-name" className="text-right text-sm">
                  Display Name
                </Label>
                <Input
                  id="display-name"
                  defaultValue="John Doe"
                  className="col-span-3"
                />
              </div>
            </div>
          </div>

          <div className="grid gap-3">
            <Label className="text-sm font-medium">Preferences</Label>
            <div className="grid gap-2">
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="language" className="text-right text-sm">
                  Language
                </Label>
                <select
                  id="language"
                  className="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus:ring-ring col-span-3 flex h-10 w-full rounded-md border px-3 py-2 text-sm focus:ring focus:ring-offset-2 focus:outline-none"
                  defaultValue="en"
                >
                  <option value="en">English</option>
                  <option value="es">Spanish</option>
                  <option value="fr">French</option>
                </select>
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="timezone" className="text-right text-sm">
                  Timezone
                </Label>
                <select
                  id="timezone"
                  className="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus:ring-ring col-span-3 flex h-10 w-full rounded-md border px-3 py-2 text-sm focus:ring focus:ring-offset-2 focus:outline-none"
                  defaultValue="utc"
                >
                  <option value="utc">UTC</option>
                  <option value="est">Eastern Time</option>
                  <option value="pst">Pacific Time</option>
                </select>
              </div>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline">Cancel</Button>
          <Button>Save Changes</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};

// User profile dialog
export const UserProfile: Story = {
  render: () => (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="ghost" size="sm">
          <User className="h-4 w-4" />
          View Profile
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <User className="h-5 w-5" />
            User Profile
          </DialogTitle>
          <DialogDescription>
            View and manage user information.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="flex items-center space-x-4">
            <div className="bg-muted flex h-12 w-12 items-center justify-center rounded-full">
              <User className="h-6 w-6" />
            </div>
            <div>
              <h3 className="font-medium">John Doe</h3>
              <p className="text-muted-foreground text-sm">
                john.doe@example.com
              </p>
            </div>
          </div>

          <div className="grid gap-2">
            <Label>Bio</Label>
            <p className="text-muted-foreground text-sm">
              A passionate developer who loves creating amazing user
              experiences. Currently working on various open-source projects and
              learning new technologies.
            </p>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label className="text-sm">Joined</Label>
              <p className="text-muted-foreground text-sm">January 2023</p>
            </div>
            <div>
              <Label className="text-sm">Location</Label>
              <p className="text-muted-foreground text-sm">San Francisco, CA</p>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline">
            <Edit className="h-4 w-4" />
            Edit Profile
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};

// Controlled dialog
export const Controlled: Story = {
  render: () => {
    const [open, setOpen] = useState(false);

    return (
      <>
        <Button onClick={() => setOpen(true)}>Open Controlled Dialog</Button>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Controlled Dialog</DialogTitle>
              <DialogDescription>
                This dialog's open state is controlled by React state.
              </DialogDescription>
            </DialogHeader>
            <div className="py-4">
              <p className="text-muted-foreground text-sm">
                You can control this dialog programmatically. The state is
                managed externally and passed to the Dialog component.
              </p>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button onClick={() => setOpen(false)}>Confirm</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </>
    );
  },
};

// Multiple dialogs
export const MultipleDialogs: Story = {
  render: () => (
    <div className="flex gap-3">
      <Dialog>
        <DialogTrigger asChild>
          <Button>Dialog 1</Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>First Dialog</DialogTitle>
            <DialogDescription>This is the first dialog.</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button>Close</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog>
        <DialogTrigger asChild>
          <Button variant="outline">Dialog 2</Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Second Dialog</DialogTitle>
            <DialogDescription>
              This is the second dialog with different content.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button>Close</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog>
        <DialogTrigger asChild>
          <Button variant="secondary">Dialog 3</Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Third Dialog</DialogTitle>
            <DialogDescription>
              Each dialog operates independently.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button>Close</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Multiple independent dialog instances.",
      },
    },
  },
};
