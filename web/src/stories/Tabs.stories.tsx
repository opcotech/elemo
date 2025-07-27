import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  BarChart3,
  Bell,
  CreditCard,
  Database,
  Download,
  Edit,
  FileText,
  Filter,
  Lock,
  Plus,
  Settings,
  Upload,
  User,
  Users,
} from "lucide-react";
import { useState } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";

const meta: Meta<typeof Tabs> = {
  title: "UI/Tabs",
  component: Tabs,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A set of layered sections of content—known as tab panels—that are displayed one at a time. Built on top of Radix UI Tabs primitive.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    value: {
      control: "text",
      description: "The controlled value of the tab to activate",
    },
    defaultValue: {
      control: "text",
      description:
        "The value of the tab that should be active when initially rendered",
    },
    onValueChange: {
      action: "onValueChange",
      description: "Callback fired when the value changes",
    },
    orientation: {
      control: "select",
      options: ["horizontal", "vertical"],
      description: "The orientation of the tabs",
    },
    dir: {
      control: "select",
      options: ["ltr", "rtl"],
      description: "The reading direction of the tabs",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic tabs
export const Default: Story = {
  render: () => (
    <Tabs defaultValue="account" className="w-[400px]">
      <TabsList>
        <TabsTrigger value="account">Account</TabsTrigger>
        <TabsTrigger value="password">Password</TabsTrigger>
      </TabsList>
      <TabsContent value="account" className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="name">Name</Label>
          <Input id="name" defaultValue="Pedro Duarte" />
        </div>
        <div className="space-y-2">
          <Label htmlFor="username">Username</Label>
          <Input id="username" defaultValue="@peduarte" />
        </div>
        <Button>Save changes</Button>
      </TabsContent>
      <TabsContent value="password" className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="current">Current password</Label>
          <Input id="current" type="password" />
        </div>
        <div className="space-y-2">
          <Label htmlFor="new">New password</Label>
          <Input id="new" type="password" />
        </div>
        <Button>Change password</Button>
      </TabsContent>
    </Tabs>
  ),
};

// With icons
export const WithIcons: Story = {
  render: () => (
    <Tabs defaultValue="profile" className="w-[400px]">
      <TabsList>
        <TabsTrigger value="profile" className="flex items-center gap-2">
          <User className="h-4 w-4" />
          Profile
        </TabsTrigger>
        <TabsTrigger value="settings" className="flex items-center gap-2">
          <Settings className="h-4 w-4" />
          Settings
        </TabsTrigger>
        <TabsTrigger value="notifications" className="flex items-center gap-2">
          <Bell className="h-4 w-4" />
          Notifications
        </TabsTrigger>
      </TabsList>
      <TabsContent value="profile" className="space-y-4">
        <div className="flex items-center space-x-4">
          <Avatar className="h-16 w-16">
            <AvatarImage src="https://github.com/shadcn.png" alt="Profile" />
            <AvatarFallback>JD</AvatarFallback>
          </Avatar>
          <div>
            <h3 className="text-lg font-medium">John Doe</h3>
            <p className="text-muted-foreground text-sm">Software Engineer</p>
          </div>
        </div>
        <div className="space-y-2">
          <Label htmlFor="bio">Bio</Label>
          <Textarea
            id="bio"
            placeholder="Tell us about yourself"
            defaultValue="I'm a passionate developer who loves creating amazing user experiences."
          />
        </div>
      </TabsContent>
      <TabsContent value="settings" className="space-y-4">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <Label htmlFor="dark-mode">Dark mode</Label>
            <Switch id="dark-mode" />
          </div>
          <div className="flex items-center justify-between">
            <Label htmlFor="auto-save">Auto-save</Label>
            <Switch id="auto-save" defaultChecked />
          </div>
          <div className="space-y-2">
            <Label htmlFor="language">Language</Label>
            <select
              id="language"
              className="border-input bg-background ring-offset-background flex h-10 w-full rounded-md border px-3 py-2 text-sm"
            >
              <option value="en">English</option>
              <option value="es">Spanish</option>
              <option value="fr">French</option>
            </select>
          </div>
        </div>
      </TabsContent>
      <TabsContent value="notifications" className="space-y-4">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <Label htmlFor="email-notifications">Email notifications</Label>
            <Switch id="email-notifications" defaultChecked />
          </div>
          <div className="flex items-center justify-between">
            <Label htmlFor="push-notifications">Push notifications</Label>
            <Switch id="push-notifications" />
          </div>
          <div className="flex items-center justify-between">
            <Label htmlFor="marketing">Marketing emails</Label>
            <Switch id="marketing" />
          </div>
        </div>
      </TabsContent>
    </Tabs>
  ),
};

// Dashboard tabs
export const DashboardTabs: Story = {
  render: () => (
    <Tabs defaultValue="overview" className="w-[600px]">
      <TabsList>
        <TabsTrigger value="overview">
          <BarChart3 className="mr-2 h-4 w-4" />
          Overview
        </TabsTrigger>
        <TabsTrigger value="analytics">
          <Database className="mr-2 h-4 w-4" />
          Analytics
        </TabsTrigger>
        <TabsTrigger value="reports">
          <FileText className="mr-2 h-4 w-4" />
          Reports
        </TabsTrigger>
        <TabsTrigger value="users">
          <Users className="mr-2 h-4 w-4" />
          Users
        </TabsTrigger>
      </TabsList>
      <TabsContent value="overview" className="space-y-4">
        <div className="grid grid-cols-3 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">
                Total Revenue
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">$45,231.89</div>
              <p className="text-muted-foreground text-xs">
                +20.1% from last month
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">
                Subscriptions
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">+2350</div>
              <p className="text-muted-foreground text-xs">
                +180.1% from last month
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Sales</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">+12,234</div>
              <p className="text-muted-foreground text-xs">
                +19% from last month
              </p>
            </CardContent>
          </Card>
        </div>
      </TabsContent>
      <TabsContent value="analytics" className="space-y-4">
        <div className="flex h-[200px] items-center justify-center rounded-lg border">
          <div className="text-center">
            <BarChart3 className="text-muted-foreground mx-auto mb-2 h-8 w-8" />
            <p className="text-muted-foreground text-sm">
              Analytics chart would go here
            </p>
          </div>
        </div>
      </TabsContent>
      <TabsContent value="reports" className="space-y-4">
        <div className="space-y-2">
          <div className="flex items-center justify-between rounded-lg border p-3">
            <div className="flex items-center space-x-3">
              <FileText className="h-5 w-5" />
              <div>
                <div className="font-medium">Monthly Report</div>
                <div className="text-muted-foreground text-sm">
                  Generated 2 hours ago
                </div>
              </div>
            </div>
            <Button variant="outline" size="sm">
              <Download className="mr-2 h-4 w-4" />
              Download
            </Button>
          </div>
          <div className="flex items-center justify-between rounded-lg border p-3">
            <div className="flex items-center space-x-3">
              <FileText className="h-5 w-5" />
              <div>
                <div className="font-medium">Quarterly Report</div>
                <div className="text-muted-foreground text-sm">
                  Generated 1 day ago
                </div>
              </div>
            </div>
            <Button variant="outline" size="sm">
              <Download className="mr-2 h-4 w-4" />
              Download
            </Button>
          </div>
        </div>
      </TabsContent>
      <TabsContent value="users" className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-medium">User Management</h3>
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            Add User
          </Button>
        </div>
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="flex items-center justify-between rounded-lg border p-3"
            >
              <div className="flex items-center space-x-3">
                <Avatar>
                  <AvatarFallback>U{i}</AvatarFallback>
                </Avatar>
                <div>
                  <div className="font-medium">User {i}</div>
                  <div className="text-muted-foreground text-sm">
                    user{i}@example.com
                  </div>
                </div>
              </div>
              <Badge variant="secondary">Active</Badge>
            </div>
          ))}
        </div>
      </TabsContent>
    </Tabs>
  ),
};

// Controlled tabs
export const ControlledTabs: Story = {
  render: () => {
    const [activeTab, setActiveTab] = useState("tab1");

    return (
      <div className="space-y-4">
        <div className="text-muted-foreground text-sm">
          Active tab: {activeTab}
        </div>
        <Tabs
          value={activeTab}
          onValueChange={setActiveTab}
          className="w-[400px]"
        >
          <TabsList>
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
            <TabsTrigger value="tab2">Tab 2</TabsTrigger>
            <TabsTrigger value="tab3">Tab 3</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">
            <Card>
              <CardHeader>
                <CardTitle>Tab 1 Content</CardTitle>
              </CardHeader>
              <CardContent>
                <p>
                  This is the content for tab 1. The active tab state is
                  controlled externally.
                </p>
              </CardContent>
            </Card>
          </TabsContent>
          <TabsContent value="tab2">
            <Card>
              <CardHeader>
                <CardTitle>Tab 2 Content</CardTitle>
              </CardHeader>
              <CardContent>
                <p>
                  This is the content for tab 2. You can programmatically change
                  tabs.
                </p>
              </CardContent>
            </Card>
          </TabsContent>
          <TabsContent value="tab3">
            <Card>
              <CardHeader>
                <CardTitle>Tab 3 Content</CardTitle>
              </CardHeader>
              <CardContent>
                <p>
                  This is the content for tab 3. The state is managed by React.
                </p>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
        <div className="flex space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setActiveTab("tab1")}
          >
            Go to Tab 1
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setActiveTab("tab2")}
          >
            Go to Tab 2
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setActiveTab("tab3")}
          >
            Go to Tab 3
          </Button>
        </div>
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "A tabs component with controlled state that can be programmatically changed.",
      },
    },
  },
};

// Content management tabs
export const ContentManagement: Story = {
  render: () => (
    <Tabs defaultValue="posts" className="w-[600px]">
      <TabsList>
        <TabsTrigger value="posts">Posts</TabsTrigger>
        <TabsTrigger value="pages">Pages</TabsTrigger>
        <TabsTrigger value="media">Media</TabsTrigger>
        <TabsTrigger value="comments">Comments</TabsTrigger>
      </TabsList>
      <TabsContent value="posts" className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-medium">Blog Posts</h3>
          <div className="flex space-x-2">
            <Button variant="outline" size="sm">
              <Filter className="mr-2 h-4 w-4" />
              Filter
            </Button>
            <Button size="sm">
              <Plus className="mr-2 h-4 w-4" />
              New Post
            </Button>
          </div>
        </div>
        <div className="space-y-2">
          {[
            {
              title: "Getting Started with React",
              status: "Published",
              date: "2023-12-01",
            },
            {
              title: "Advanced TypeScript Tips",
              status: "Draft",
              date: "2023-11-28",
            },
            {
              title: "Building Better UIs",
              status: "Published",
              date: "2023-11-25",
            },
          ].map((post, i) => (
            <div
              key={i}
              className="flex items-center justify-between rounded-lg border p-3"
            >
              <div>
                <div className="font-medium">{post.title}</div>
                <div className="text-muted-foreground text-sm">{post.date}</div>
              </div>
              <div className="flex items-center space-x-2">
                <Badge
                  variant={
                    post.status === "Published" ? "default" : "secondary"
                  }
                >
                  {post.status}
                </Badge>
                <Button variant="ghost" size="sm">
                  <Edit className="h-4 w-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      </TabsContent>
      <TabsContent value="pages" className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-medium">Pages</h3>
          <Button size="sm">
            <Plus className="mr-2 h-4 w-4" />
            New Page
          </Button>
        </div>
        <div className="space-y-2">
          {[
            { title: "About Us", status: "Published" },
            { title: "Contact", status: "Published" },
            { title: "Privacy Policy", status: "Draft" },
          ].map((page, i) => (
            <div
              key={i}
              className="flex items-center justify-between rounded-lg border p-3"
            >
              <div className="font-medium">{page.title}</div>
              <div className="flex items-center space-x-2">
                <Badge
                  variant={
                    page.status === "Published" ? "default" : "secondary"
                  }
                >
                  {page.status}
                </Badge>
                <Button variant="ghost" size="sm">
                  <Edit className="h-4 w-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      </TabsContent>
      <TabsContent value="media" className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-medium">Media Library</h3>
          <Button size="sm">
            <Upload className="mr-2 h-4 w-4" />
            Upload
          </Button>
        </div>
        <div className="grid grid-cols-4 gap-4">
          {Array.from({ length: 8 }, (_, i) => (
            <div
              key={i}
              className="bg-muted flex aspect-square items-center justify-center rounded-lg border"
            >
              <div className="text-center">
                <div className="mb-1 text-2xl">🖼️</div>
                <div className="text-muted-foreground text-xs">
                  Image {i + 1}
                </div>
              </div>
            </div>
          ))}
        </div>
      </TabsContent>
      <TabsContent value="comments" className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-medium">Comments</h3>
          <div className="flex space-x-2">
            <Button variant="outline" size="sm">
              <Filter className="mr-2 h-4 w-4" />
              Filter
            </Button>
          </div>
        </div>
        <div className="space-y-2">
          {[
            {
              author: "John Doe",
              comment: "Great article! Very helpful.",
              status: "Approved",
            },
            {
              author: "Jane Smith",
              comment: "Thanks for sharing this.",
              status: "Pending",
            },
            {
              author: "Bob Wilson",
              comment: "Could you elaborate on this point?",
              status: "Approved",
            },
          ].map((comment, i) => (
            <div key={i} className="space-y-2 rounded-lg border p-3">
              <div className="flex items-center justify-between">
                <div className="font-medium">{comment.author}</div>
                <Badge
                  variant={
                    comment.status === "Approved" ? "default" : "secondary"
                  }
                >
                  {comment.status}
                </Badge>
              </div>
              <p className="text-sm">{comment.comment}</p>
            </div>
          ))}
        </div>
      </TabsContent>
    </Tabs>
  ),
};

// Vertical tabs
export const VerticalTabs: Story = {
  render: () => (
    <Tabs
      defaultValue="general"
      orientation="vertical"
      className="flex w-[600px]"
    >
      <TabsList className="h-auto w-[200px] flex-col p-1">
        <TabsTrigger value="general" className="w-full justify-start">
          <Settings className="mr-2 h-4 w-4" />
          General
        </TabsTrigger>
        <TabsTrigger value="security" className="w-full justify-start">
          <Lock className="mr-2 h-4 w-4" />
          Security
        </TabsTrigger>
        <TabsTrigger value="billing" className="w-full justify-start">
          <CreditCard className="mr-2 h-4 w-4" />
          Billing
        </TabsTrigger>
        <TabsTrigger value="notifications" className="w-full justify-start">
          <Bell className="mr-2 h-4 w-4" />
          Notifications
        </TabsTrigger>
      </TabsList>
      <div className="ml-4 flex-1">
        <TabsContent value="general" className="mt-0 space-y-4">
          <div className="space-y-2">
            <h3 className="text-lg font-medium">General Settings</h3>
            <p className="text-muted-foreground text-sm">
              Manage your general account settings and preferences.
            </p>
          </div>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input id="name" defaultValue="John Doe" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input id="email" type="email" defaultValue="john@example.com" />
            </div>
          </div>
        </TabsContent>
        <TabsContent value="security" className="mt-0 space-y-4">
          <div className="space-y-2">
            <h3 className="text-lg font-medium">Security Settings</h3>
            <p className="text-muted-foreground text-sm">
              Manage your account security and authentication.
            </p>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <Label htmlFor="2fa">Two-factor authentication</Label>
              <Switch id="2fa" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="current-password">Current Password</Label>
              <Input id="current-password" type="password" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="new-password">New Password</Label>
              <Input id="new-password" type="password" />
            </div>
          </div>
        </TabsContent>
        <TabsContent value="billing" className="mt-0 space-y-4">
          <div className="space-y-2">
            <h3 className="text-lg font-medium">Billing Settings</h3>
            <p className="text-muted-foreground text-sm">
              Manage your subscription and payment methods.
            </p>
          </div>
          <div className="space-y-4">
            <div className="rounded-lg border p-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="font-medium">Pro Plan</div>
                  <div className="text-muted-foreground text-sm">$19/month</div>
                </div>
                <Badge variant="default">Active</Badge>
              </div>
            </div>
            <Button>Manage Subscription</Button>
          </div>
        </TabsContent>
        <TabsContent value="notifications" className="mt-0 space-y-4">
          <div className="space-y-2">
            <h3 className="text-lg font-medium">Notification Settings</h3>
            <p className="text-muted-foreground text-sm">
              Configure how you receive notifications.
            </p>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <Label htmlFor="email-notifications">Email notifications</Label>
              <Switch id="email-notifications" defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="push-notifications">Push notifications</Label>
              <Switch id="push-notifications" />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="sms-notifications">SMS notifications</Label>
              <Switch id="sms-notifications" />
            </div>
          </div>
        </TabsContent>
      </div>
    </Tabs>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Tabs arranged vertically with a sidebar-like navigation layout.",
      },
    },
  },
};

// Many tabs
export const ManyTabs: Story = {
  render: () => (
    <Tabs defaultValue="tab1" className="w-[600px]">
      <TabsList className="grid w-full grid-cols-6">
        <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        <TabsTrigger value="tab2">Tab 2</TabsTrigger>
        <TabsTrigger value="tab3">Tab 3</TabsTrigger>
        <TabsTrigger value="tab4">Tab 4</TabsTrigger>
        <TabsTrigger value="tab5">Tab 5</TabsTrigger>
        <TabsTrigger value="tab6">Tab 6</TabsTrigger>
      </TabsList>
      {Array.from({ length: 6 }, (_, i) => (
        <TabsContent key={i} value={`tab${i + 1}`}>
          <Card>
            <CardHeader>
              <CardTitle>Tab {i + 1} Content</CardTitle>
              <CardDescription>
                This is the content for tab {i + 1}.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p>
                Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do
                eiusmod tempor incididunt ut labore et dolore magna aliqua. Tab{" "}
                {i + 1} specific content here.
              </p>
            </CardContent>
          </Card>
        </TabsContent>
      ))}
    </Tabs>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "A tabs component with many tabs using a grid layout for the tab list.",
      },
    },
  },
};
