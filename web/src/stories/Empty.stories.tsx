import type { Meta, StoryObj } from "@storybook/react-vite";
import { FolderOpen, Inbox, Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

const meta: Meta<typeof Empty> = {
  title: "UI/Empty",
  component: Empty,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Empty state placeholder for lists, tables, and panels with no content.",
      },
    },
  },
  tags: ["autodocs"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Empty className="w-[400px] border">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Inbox />
        </EmptyMedia>
        <EmptyTitle>No messages</EmptyTitle>
        <EmptyDescription>
          You do not have any messages yet. Start a conversation to see it here.
        </EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button>
          <Plus className="h-4 w-4" />
          New message
        </Button>
      </EmptyContent>
    </Empty>
  ),
};

export const WithFolder: Story = {
  render: () => (
    <Empty className="w-[400px] border">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <FolderOpen />
        </EmptyMedia>
        <EmptyTitle>No projects</EmptyTitle>
        <EmptyDescription>
          Create your first project to get started.
        </EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button variant="outline">Create project</Button>
      </EmptyContent>
    </Empty>
  ),
};
