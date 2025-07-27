import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  File,
  Folder,
  Image,
  MoreHorizontal,
  Music,
  Video,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";

const meta: Meta<typeof ScrollArea> = {
  title: "UI/ScrollArea",
  component: ScrollArea,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A custom scrollable area with a more visually appealing scrollbar. Built on top of Radix UI ScrollArea primitive.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    className: {
      control: "text",
      description: "Additional CSS classes to apply to the scroll area",
    },
    type: {
      control: "select",
      options: ["auto", "always", "scroll", "hover"],
      description: "When to show scrollbars",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic scroll area
export const Default: Story = {
  render: () => (
    <ScrollArea className="h-72 w-48 rounded-md border">
      <div className="p-4">
        <h4 className="mb-4 text-sm leading-none font-medium">Tags</h4>
        {Array.from({ length: 50 }, (_, i) => (
          <div key={i} className="text-sm">
            v1.2.{i}
          </div>
        ))}
      </div>
    </ScrollArea>
  ),
};

// Horizontal scroll
export const Horizontal: Story = {
  render: () => (
    <ScrollArea className="w-96 rounded-md border whitespace-nowrap">
      <div className="flex w-max space-x-4 p-4">
        {Array.from({ length: 20 }, (_, i) => (
          <div
            key={i}
            className="bg-muted flex h-20 w-20 items-center justify-center rounded-md text-sm font-medium"
          >
            {i + 1}
          </div>
        ))}
      </div>
    </ScrollArea>
  ),
};

// File list
export const FileList: Story = {
  render: () => {
    const files = [
      { name: "Documents", type: "folder", size: "", modified: "2 hours ago" },
      { name: "Images", type: "folder", size: "", modified: "1 day ago" },
      { name: "Videos", type: "folder", size: "", modified: "3 days ago" },
      {
        name: "report.pdf",
        type: "pdf",
        size: "2.1 MB",
        modified: "5 minutes ago",
      },
      {
        name: "presentation.pptx",
        type: "presentation",
        size: "15.8 MB",
        modified: "1 hour ago",
      },
      {
        name: "data.xlsx",
        type: "spreadsheet",
        size: "890 KB",
        modified: "2 hours ago",
      },
      {
        name: "photo1.jpg",
        type: "image",
        size: "3.2 MB",
        modified: "1 day ago",
      },
      {
        name: "photo2.jpg",
        type: "image",
        size: "2.8 MB",
        modified: "1 day ago",
      },
      {
        name: "song.mp3",
        type: "audio",
        size: "4.5 MB",
        modified: "2 days ago",
      },
      {
        name: "movie.mp4",
        type: "video",
        size: "125 MB",
        modified: "3 days ago",
      },
      {
        name: "archive.zip",
        type: "archive",
        size: "45 MB",
        modified: "1 week ago",
      },
      {
        name: "readme.txt",
        type: "text",
        size: "1.2 KB",
        modified: "2 weeks ago",
      },
      {
        name: "config.json",
        type: "code",
        size: "850 B",
        modified: "1 month ago",
      },
      {
        name: "backup_old.tar.gz",
        type: "archive",
        size: "2.1 GB",
        modified: "3 months ago",
      },
    ];

    const getIcon = (type: string) => {
      switch (type) {
        case "folder":
          return <Folder className="h-4 w-4 text-blue-500" />;
        case "image":
          return <Image className="h-4 w-4 text-green-500" />;
        case "audio":
          return <Music className="h-4 w-4 text-purple-500" />;
        case "video":
          return <Video className="h-4 w-4 text-red-500" />;
        default:
          return <File className="h-4 w-4 text-gray-500" />;
      }
    };

    return (
      <ScrollArea className="h-80 w-96 rounded-md border">
        <div className="p-4">
          <h4 className="mb-4 text-sm leading-none font-medium">Files</h4>
          <div className="space-y-1">
            {files.map((file, index) => (
              <div
                key={index}
                className="hover:bg-muted flex cursor-pointer items-center justify-between rounded-md p-2"
              >
                <div className="flex items-center space-x-2">
                  {getIcon(file.type)}
                  <div>
                    <div className="text-sm font-medium">{file.name}</div>
                    <div className="text-muted-foreground text-xs">
                      {file.size && `${file.size} • `}
                      {file.modified}
                    </div>
                  </div>
                </div>
                <Button variant="ghost" size="sm">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>
        </div>
      </ScrollArea>
    );
  },
};

// Chat messages
export const ChatMessages: Story = {
  render: () => {
    const messages = [
      {
        id: 1,
        user: "Alice",
        message: "Hey everyone! How's it going?",
        time: "10:30 AM",
        own: false,
      },
      {
        id: 2,
        user: "You",
        message: "Hi Alice! Doing well, thanks for asking.",
        time: "10:32 AM",
        own: true,
      },
      {
        id: 3,
        user: "Bob",
        message: "Great! Just finished the project presentation.",
        time: "10:35 AM",
        own: false,
      },
      {
        id: 4,
        user: "Charlie",
        message: "Awesome work Bob! 🎉",
        time: "10:36 AM",
        own: false,
      },
      {
        id: 5,
        user: "You",
        message: "Congratulations! How did it go?",
        time: "10:37 AM",
        own: true,
      },
      {
        id: 6,
        user: "Bob",
        message: "Really well! The client loved the design.",
        time: "10:40 AM",
        own: false,
      },
      {
        id: 7,
        user: "Alice",
        message: "That's fantastic news! 🚀",
        time: "10:42 AM",
        own: false,
      },
      {
        id: 8,
        user: "Dave",
        message: "Sorry I'm late to the conversation. Congrats Bob!",
        time: "10:45 AM",
        own: false,
      },
      {
        id: 9,
        user: "You",
        message: "We should celebrate this weekend!",
        time: "10:47 AM",
        own: true,
      },
      {
        id: 10,
        user: "Charlie",
        message: "I'm in! Let's plan something fun.",
        time: "10:48 AM",
        own: false,
      },
    ];

    return (
      <ScrollArea className="h-80 w-80 rounded-md border">
        <div className="space-y-3 p-4">
          {messages.map((msg) => (
            <div
              key={msg.id}
              className={`flex ${msg.own ? "justify-end" : "justify-start"}`}
            >
              <div
                className={`max-w-[70%] rounded-lg p-3 ${
                  msg.own ? "bg-primary text-primary-foreground" : "bg-muted"
                }`}
              >
                {!msg.own && (
                  <div className="mb-1 text-xs font-medium">{msg.user}</div>
                )}
                <div className="text-sm">{msg.message}</div>
                <div
                  className={`mt-1 text-xs ${msg.own ? "text-primary-foreground/70" : "text-muted-foreground"}`}
                >
                  {msg.time}
                </div>
              </div>
            </div>
          ))}
        </div>
      </ScrollArea>
    );
  },
};

// Notifications list
export const NotificationsList: Story = {
  render: () => {
    const notifications = [
      {
        id: 1,
        title: "New message from Sarah",
        description: "Hey! Are you available for a quick call?",
        time: "2 min ago",
        unread: true,
        type: "message",
      },
      {
        id: 2,
        title: "Project deadline reminder",
        description: "The project is due in 2 days. Please review your tasks.",
        time: "1 hour ago",
        unread: true,
        type: "reminder",
      },
      {
        id: 3,
        title: "System update completed",
        description:
          "Your system has been successfully updated to version 2.1.0",
        time: "3 hours ago",
        unread: false,
        type: "system",
      },
      {
        id: 4,
        title: "New team member joined",
        description: "Alex Johnson has joined the development team.",
        time: "5 hours ago",
        unread: false,
        type: "team",
      },
      {
        id: 5,
        title: "Weekly report ready",
        description: "Your weekly performance report is now available.",
        time: "1 day ago",
        unread: false,
        type: "report",
      },
      {
        id: 6,
        title: "Payment processed",
        description: "Your monthly subscription payment has been processed.",
        time: "2 days ago",
        unread: false,
        type: "payment",
      },
      {
        id: 7,
        title: "Security alert",
        description: "New login detected from Chrome on Windows.",
        time: "3 days ago",
        unread: false,
        type: "security",
      },
    ];

    return (
      <ScrollArea className="h-80 w-96 rounded-md border">
        <div className="p-4">
          <h4 className="mb-4 text-sm leading-none font-medium">
            Notifications
          </h4>
          <div className="space-y-1">
            {notifications.map((notification, index) => (
              <div key={notification.id}>
                <div
                  className={`hover:bg-muted cursor-pointer rounded-md p-3 ${
                    notification.unread ? "bg-muted/50" : ""
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <h5 className="text-sm font-medium">
                          {notification.title}
                        </h5>
                        {notification.unread && (
                          <div className="h-2 w-2 rounded-full bg-blue-500"></div>
                        )}
                      </div>
                      <p className="text-muted-foreground mt-1 text-xs">
                        {notification.description}
                      </p>
                      <div className="text-muted-foreground mt-2 text-xs">
                        {notification.time}
                      </div>
                    </div>
                  </div>
                </div>
                {index < notifications.length - 1 && (
                  <Separator className="my-1" />
                )}
              </div>
            ))}
          </div>
        </div>
      </ScrollArea>
    );
  },
};

// Code editor
export const CodeEditor: Story = {
  render: () => {
    const code = `import React from 'react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';

interface LoginFormProps {
  onSubmit: (data: { email: string; password: string }) => void;
}

export function LoginForm({ onSubmit }: LoginFormProps) {
  const [email, setEmail] = React.useState('');
  const [password, setPassword] = React.useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({ email, password });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="email">Email</Label>
        <Input
          id="email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="Enter your email"
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="password">Password</Label>
        <Input
          id="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Enter your password"
          required
        />
      </div>

      <Button type="submit" className="w-full">
        Sign In
      </Button>
    </form>
  );
}`;

    return (
      <ScrollArea className="h-80 w-96 rounded-md border">
        <div className="p-4">
          <div className="mb-4 flex items-center justify-between">
            <h4 className="text-sm font-medium">LoginForm.tsx</h4>
            <Badge variant="secondary">TypeScript</Badge>
          </div>
          <pre className="font-mono text-xs">
            <code>{code}</code>
          </pre>
        </div>
      </ScrollArea>
    );
  },
};

// Both scrollbars
export const BothScrollbars: Story = {
  render: () => (
    <ScrollArea className="h-60 w-80 rounded-md border">
      <div className="p-4" style={{ width: "600px" }}>
        <h4 className="mb-4 text-sm leading-none font-medium">
          Wide Content with Both Scrollbars
        </h4>
        {Array.from({ length: 30 }, (_, i) => (
          <div key={i} className="mb-2 text-sm">
            This is a very long line of text that extends beyond the container
            width to demonstrate horizontal scrolling. Line {i + 1}
          </div>
        ))}
      </div>
    </ScrollArea>
  ),
  parameters: {
    docs: {
      description: {
        story: "ScrollArea with both vertical and horizontal scrollbars.",
      },
    },
  },
};

// Custom styling
export const CustomStyling: Story = {
  render: () => (
    <ScrollArea className="border-primary/50 bg-muted/30 h-72 w-48 rounded-lg border-2 border-dashed">
      <div className="p-6">
        <h4 className="text-primary mb-4 text-sm leading-none font-medium">
          Custom Styled
        </h4>
        {Array.from({ length: 25 }, (_, i) => (
          <div key={i} className="bg-background mb-2 rounded p-2 text-sm">
            Custom item {i + 1}
          </div>
        ))}
      </div>
    </ScrollArea>
  ),
  parameters: {
    docs: {
      description: {
        story: "ScrollArea with custom border and background styling.",
      },
    },
  },
};
