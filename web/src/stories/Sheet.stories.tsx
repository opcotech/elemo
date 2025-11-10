import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  Bell,
  Download,
  Edit,
  Filter,
  Mail,
  MapPin,
  Menu,
  Phone,
  Plus,
  Search,
  Settings,
  Share,
  Star,
  Trash2,
  Upload,
  User,
} from "lucide-react";
import { useState } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";

const meta: Meta<typeof Sheet> = {
  title: "UI/Sheet",
  component: Sheet,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Extends the Dialog component to display content that complements the main content of the screen. Built on top of Radix UI Dialog primitive.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    open: {
      control: "boolean",
      description: "Controls the open state of the sheet",
    },
    onOpenChange: {
      action: "onOpenChange",
      description: "Callback fired when the open state changes",
    },
    modal: {
      control: "boolean",
      description: "Whether the sheet is modal",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic sheet from right
export const Default: Story = {
  render: () => (
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="outline">Open Sheet</Button>
      </SheetTrigger>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>Edit profile</SheetTitle>
          <SheetDescription>
            Make changes to your profile here. Click save when you're done.
          </SheetDescription>
        </SheetHeader>
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
        <SheetFooter>
          <SheetClose asChild>
            <Button type="submit">Save changes</Button>
          </SheetClose>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  ),
};

// Different sides
export const DifferentSides: Story = {
  render: () => (
    <div className="flex gap-4">
      <Sheet>
        <SheetTrigger asChild>
          <Button variant="outline">Left</Button>
        </SheetTrigger>
        <SheetContent side="left">
          <SheetHeader>
            <SheetTitle>Left Sheet</SheetTitle>
            <SheetDescription>
              This sheet opens from the left side.
            </SheetDescription>
          </SheetHeader>
        </SheetContent>
      </Sheet>

      <Sheet>
        <SheetTrigger asChild>
          <Button variant="outline">Right</Button>
        </SheetTrigger>
        <SheetContent side="right">
          <SheetHeader>
            <SheetTitle>Right Sheet</SheetTitle>
            <SheetDescription>
              This sheet opens from the right side.
            </SheetDescription>
          </SheetHeader>
        </SheetContent>
      </Sheet>

      <Sheet>
        <SheetTrigger asChild>
          <Button variant="outline">Top</Button>
        </SheetTrigger>
        <SheetContent side="top">
          <SheetHeader>
            <SheetTitle>Top Sheet</SheetTitle>
            <SheetDescription>This sheet opens from the top.</SheetDescription>
          </SheetHeader>
        </SheetContent>
      </Sheet>

      <Sheet>
        <SheetTrigger asChild>
          <Button variant="outline">Bottom</Button>
        </SheetTrigger>
        <SheetContent side="bottom">
          <SheetHeader>
            <SheetTitle>Bottom Sheet</SheetTitle>
            <SheetDescription>
              This sheet opens from the bottom.
            </SheetDescription>
          </SheetHeader>
        </SheetContent>
      </Sheet>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Sheets that can open from different sides of the screen.",
      },
    },
  },
};

// Navigation menu
export const NavigationMenu: Story = {
  render: () => (
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="outline" size="icon">
          <Menu className="h-4 w-4" />
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="w-80">
        <SheetHeader>
          <SheetTitle>Navigation</SheetTitle>
        </SheetHeader>
        <nav className="mt-6 space-y-4">
          <div className="space-y-2">
            <h4 className="text-sm font-medium">Main</h4>
            <div className="space-y-1">
              <Button variant="ghost" className="w-full justify-start">
                <User className="size-4" />
                Profile
              </Button>
              <Button variant="ghost" className="w-full justify-start">
                <Settings className="size-4" />
                Settings
              </Button>
              <Button variant="ghost" className="w-full justify-start">
                <Bell className="size-4" />
                Notifications
              </Button>
            </div>
          </div>

          <Separator />

          <div className="space-y-2">
            <h4 className="text-sm font-medium">Tools</h4>
            <div className="space-y-1">
              <Button variant="ghost" className="w-full justify-start">
                <Search className="size-4" />
                Search
              </Button>
              <Button variant="ghost" className="w-full justify-start">
                <Filter className="size-4" />
                Filters
              </Button>
            </div>
          </div>

          <Separator />

          <div className="space-y-2">
            <h4 className="text-sm font-medium">Account</h4>
            <div className="space-y-1">
              <Button variant="ghost" className="w-full justify-start">
                <Edit className="size-4" />
                Edit Profile
              </Button>
              <Button
                variant="ghost"
                className="w-full justify-start text-red-600"
              >
                <Trash2 className="size-4" />
                Delete Account
              </Button>
            </div>
          </div>
        </nav>
      </SheetContent>
    </Sheet>
  ),
};

// Add new item form
export const AddItemForm: Story = {
  render: () => (
    <Sheet>
      <SheetTrigger asChild>
        <Button>
          <Plus className="size-4" />
          Add Item
        </Button>
      </SheetTrigger>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>Add New Item</SheetTitle>
          <SheetDescription>
            Create a new item by filling out the form below.
          </SheetDescription>
        </SheetHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="item-name">Name</Label>
            <Input id="item-name" placeholder="Enter item name" />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              placeholder="Enter item description"
              rows={3}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="category">Category</Label>
            <select
              id="category"
              className="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus:ring-ring flex h-10 w-full rounded-md border px-3 py-2 text-sm focus:ring focus:ring-offset-2 focus:outline-none"
            >
              <option value="">Select category</option>
              <option value="electronics">Electronics</option>
              <option value="clothing">Clothing</option>
              <option value="books">Books</option>
              <option value="home">Home & Garden</option>
            </select>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="price">Price</Label>
            <Input id="price" type="number" placeholder="0.00" />
          </div>
          <div className="flex items-center space-x-2">
            <Switch id="featured" />
            <Label htmlFor="featured">Featured item</Label>
          </div>
        </div>
        <SheetFooter>
          <SheetClose asChild>
            <Button variant="outline">Cancel</Button>
          </SheetClose>
          <SheetClose asChild>
            <Button type="submit">Save Item</Button>
          </SheetClose>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  ),
};

// User profile
export const UserProfile: Story = {
  render: () => (
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="ghost" size="sm">
          <Avatar className="mr-2 h-6 w-6">
            <AvatarImage src="https://github.com/shadcn.png" alt="User" />
            <AvatarFallback>JD</AvatarFallback>
          </Avatar>
          Profile
        </Button>
      </SheetTrigger>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>User Profile</SheetTitle>
          <SheetDescription>
            View and manage your profile information.
          </SheetDescription>
        </SheetHeader>
        <div className="space-y-6 py-4">
          <div className="flex items-center space-x-4">
            <Avatar className="h-16 w-16">
              <AvatarImage src="https://github.com/shadcn.png" alt="Profile" />
              <AvatarFallback>JD</AvatarFallback>
            </Avatar>
            <div>
              <h3 className="font-medium">John Doe</h3>
              <p className="text-muted-foreground text-sm">Software Engineer</p>
              <div className="mt-1 flex items-center gap-2">
                <Badge variant="secondary">Pro Member</Badge>
                <Badge variant="outline">Verified</Badge>
              </div>
            </div>
          </div>

          <Separator />

          <div className="space-y-4">
            <div className="grid gap-2">
              <Label className="text-sm font-medium">Contact Information</Label>
              <div className="space-y-2">
                <div className="flex items-center text-sm">
                  <Mail className="size-4" />
                  john.doe@example.com
                </div>
                <div className="flex items-center text-sm">
                  <Phone className="size-4" />
                  +1 (555) 123-4567
                </div>
                <div className="flex items-center text-sm">
                  <MapPin className="size-4" />
                  San Francisco, CA
                </div>
              </div>
            </div>

            <Separator />

            <div className="space-y-2">
              <Label className="text-sm font-medium">Account Stats</Label>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <div className="font-medium">142</div>
                  <div className="text-muted-foreground">Posts</div>
                </div>
                <div>
                  <div className="font-medium">1.2K</div>
                  <div className="text-muted-foreground">Followers</div>
                </div>
                <div>
                  <div className="font-medium">342</div>
                  <div className="text-muted-foreground">Following</div>
                </div>
                <div>
                  <div className="font-medium">4.8</div>
                  <div className="text-muted-foreground">Rating</div>
                </div>
              </div>
            </div>

            <Separator />

            <div className="space-y-2">
              <Label className="text-sm font-medium">Quick Actions</Label>
              <div className="grid gap-2">
                <Button variant="outline" className="justify-start">
                  <Edit className="size-4" />
                  Edit Profile
                </Button>
                <Button variant="outline" className="justify-start">
                  <Settings className="size-4" />
                  Account Settings
                </Button>
                <Button variant="outline" className="justify-start">
                  <Share className="size-4" />
                  Share Profile
                </Button>
              </div>
            </div>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  ),
};

// Filter panel
export const FilterPanel: Story = {
  render: () => (
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="outline">
          <Filter className="size-4" />
          Filters
        </Button>
      </SheetTrigger>
      <SheetContent side="left">
        <SheetHeader>
          <SheetTitle>Filter Options</SheetTitle>
          <SheetDescription>
            Customize your search and filter results.
          </SheetDescription>
        </SheetHeader>
        <div className="space-y-6 py-4">
          <div className="space-y-3">
            <Label className="text-sm font-medium">Price Range</Label>
            <div className="grid grid-cols-2 gap-2">
              <Input placeholder="Min" />
              <Input placeholder="Max" />
            </div>
          </div>

          <Separator />

          <div className="space-y-3">
            <Label className="text-sm font-medium">Category</Label>
            <div className="space-y-2">
              {[
                "Electronics",
                "Clothing",
                "Books",
                "Home & Garden",
                "Sports",
              ].map((category) => (
                <div key={category} className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    id={category}
                    className="border-input h-4 w-4 rounded border"
                  />
                  <Label htmlFor={category} className="text-sm">
                    {category}
                  </Label>
                </div>
              ))}
            </div>
          </div>

          <Separator />

          <div className="space-y-3">
            <Label className="text-sm font-medium">Rating</Label>
            <div className="space-y-2">
              {[5, 4, 3, 2, 1].map((rating) => (
                <div key={rating} className="flex items-center space-x-2">
                  <input
                    type="radio"
                    id={`rating-${rating}`}
                    name="rating"
                    className="h-4 w-4"
                  />
                  <Label
                    htmlFor={`rating-${rating}`}
                    className="flex items-center text-sm"
                  >
                    {Array.from({ length: rating }, (_, i) => (
                      <Star
                        key={i}
                        className="h-3 w-3 fill-yellow-400 text-yellow-400"
                      />
                    ))}
                    <span className="ml-1">& up</span>
                  </Label>
                </div>
              ))}
            </div>
          </div>

          <Separator />

          <div className="space-y-3">
            <Label className="text-sm font-medium">Features</Label>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="free-shipping" className="text-sm">
                  Free Shipping
                </Label>
                <Switch id="free-shipping" />
              </div>
              <div className="flex items-center justify-between">
                <Label htmlFor="on-sale" className="text-sm">
                  On Sale
                </Label>
                <Switch id="on-sale" />
              </div>
              <div className="flex items-center justify-between">
                <Label htmlFor="in-stock" className="text-sm">
                  In Stock
                </Label>
                <Switch id="in-stock" />
              </div>
            </div>
          </div>
        </div>
        <SheetFooter>
          <SheetClose asChild>
            <Button variant="outline">Clear All</Button>
          </SheetClose>
          <SheetClose asChild>
            <Button>Apply Filters</Button>
          </SheetClose>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  ),
};

// Controlled sheet
export const Controlled: Story = {
  render: () => {
    const [open, setOpen] = useState(false);

    return (
      <div className="space-y-2">
        <div className="text-muted-foreground text-sm">
          Sheet is {open ? "open" : "closed"}
        </div>
        <Sheet open={open} onOpenChange={setOpen}>
          <SheetTrigger asChild>
            <Button>Open Controlled Sheet</Button>
          </SheetTrigger>
          <SheetContent>
            <SheetHeader>
              <SheetTitle>Controlled Sheet</SheetTitle>
              <SheetDescription>
                This sheet's open state is controlled by React state.
              </SheetDescription>
            </SheetHeader>
            <div className="py-4">
              <p className="text-muted-foreground text-sm">
                You can control this sheet programmatically. The state is
                managed externally and passed to the Sheet component.
              </p>
            </div>
            <SheetFooter>
              <Button onClick={() => setOpen(false)}>Close Sheet</Button>
            </SheetFooter>
          </SheetContent>
        </Sheet>
      </div>
    );
  },
};

// File upload sheet
export const FileUploadSheet: Story = {
  render: () => (
    <Sheet>
      <SheetTrigger asChild>
        <Button>
          <Upload className="size-4" />
          Upload Files
        </Button>
      </SheetTrigger>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>Upload Files</SheetTitle>
          <SheetDescription>
            Upload your files to the cloud storage.
          </SheetDescription>
        </SheetHeader>
        <div className="space-y-6 py-4">
          <div className="border-muted-foreground rounded-lg border-2 border-dashed p-8 text-center">
            <Upload className="text-muted-foreground mx-auto mb-4 h-8 w-8" />
            <div className="text-sm">
              <span className="font-medium">Click to upload</span> or drag and
              drop
            </div>
            <div className="text-muted-foreground mt-1 text-xs">
              PNG, JPG, PDF up to 10MB
            </div>
          </div>

          <div className="space-y-3">
            <Label className="text-sm font-medium">Recent Uploads</Label>
            <div className="space-y-2">
              {[
                { name: "document.pdf", size: "2.1 MB", time: "2 minutes ago" },
                { name: "image.jpg", size: "3.4 MB", time: "1 hour ago" },
                {
                  name: "presentation.pptx",
                  size: "15.2 MB",
                  time: "Yesterday",
                },
              ].map((file, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between rounded border p-2"
                >
                  <div>
                    <div className="text-sm font-medium">{file.name}</div>
                    <div className="text-muted-foreground text-xs">
                      {file.size} • {file.time}
                    </div>
                  </div>
                  <Button variant="ghost" size="sm">
                    <Download className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          </div>
        </div>
        <SheetFooter>
          <SheetClose asChild>
            <Button variant="outline">Cancel</Button>
          </SheetClose>
          <SheetClose asChild>
            <Button>Upload Selected</Button>
          </SheetClose>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  ),
};
