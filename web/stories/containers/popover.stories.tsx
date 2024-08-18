import React from 'react';
import type { Meta, StoryObj } from '@storybook/react';

import { Button } from '@/components/ui/button';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';

const meta: Meta<typeof Popover> = {
  title: 'Containers/Popover',
  component: Popover,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  subcomponents: {
    PopoverContent: PopoverContent as any,
    PopoverTrigger: PopoverTrigger as any,
  },
  render: ({ open, ...props }) => (
    <Popover open={open} {...props}>
      <PopoverTrigger asChild>
        <Button>Open Dialog</Button>
      </PopoverTrigger>
      <PopoverContent>
        <h4>Example</h4>
        <p className='small'>Hello from the popover</p>
      </PopoverContent>
    </Popover>
  ),
};

export default meta;
type Story = StoryObj<typeof Popover>;

export const Default: Story = {};
