import type { Meta, StoryObj } from '@storybook/react';

import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

const meta: Meta<typeof DropdownMenu> = {
  title: 'Elements/Dropdown Menu',
  component: DropdownMenu,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  render: (props) => {
    return (
      <DropdownMenu {...props}>
        <DropdownMenuTrigger asChild>
          <Button>Open Dropdown</Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuLabel>My Account</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem>Profile</DropdownMenuItem>
          <DropdownMenuItem>Billing</DropdownMenuItem>
          <DropdownMenuItem>Team</DropdownMenuItem>
          <DropdownMenuItem>Subscription</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    );
  },
};

export default meta;
type Story = StoryObj<typeof DropdownMenu>;

export const Default: Story = {};

export const OpenState: Story = {
  args: {
    open: true,
    modal: false,
  },
};
