import type { Meta, StoryObj } from "@storybook/react-vite";
import { FileText, MoreHorizontal, User } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemMedia,
  ItemSeparator,
  ItemTitle,
} from "@/components/ui/item";

const meta: Meta<typeof Item> = {
  title: "UI/Item",
  component: Item,
  parameters: {
    a11y: { test: "error" },
    layout: "centered",
    docs: {
      description: {
        component:
          "Flexible list row component for menus, settings, and entity lists. AppList opts into list/listitem semantics; plain ItemGroup stays presentational so separators and mixed children remain valid.",
      },
    },
  },
  tags: ["autodocs", "verification"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <ItemGroup className="w-105">
      <Item>
        <ItemMedia variant="icon">
          <User />
        </ItemMedia>
        <ItemContent>
          <ItemTitle>Jane Cooper</ItemTitle>
          <ItemDescription>Product designer</ItemDescription>
        </ItemContent>
        <ItemActions>
          <Button variant="ghost" size="icon-sm" aria-label="More actions">
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </ItemActions>
      </Item>
      <ItemSeparator />
      <Item variant="outline">
        <ItemMedia variant="icon">
          <FileText />
        </ItemMedia>
        <ItemContent>
          <ItemTitle>Quarterly report</ItemTitle>
          <ItemDescription>Updated 2 hours ago</ItemDescription>
        </ItemContent>
      </Item>
    </ItemGroup>
  ),
};

export const Sizes: Story = {
  render: () => (
    <ItemGroup className="w-105" data-size="sm">
      <Item size="sm">
        <ItemMedia variant="icon">
          <User />
        </ItemMedia>
        <ItemContent>
          <ItemTitle>Small item</ItemTitle>
          <ItemDescription>Compact list row</ItemDescription>
        </ItemContent>
      </Item>
      <Item>
        <ItemMedia variant="icon">
          <User />
        </ItemMedia>
        <ItemContent>
          <ItemTitle>Default item</ItemTitle>
          <ItemDescription>Standard list row</ItemDescription>
        </ItemContent>
      </Item>
    </ItemGroup>
  ),
};
