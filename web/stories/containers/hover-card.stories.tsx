import React from 'react';
import type { Meta, StoryObj } from '@storybook/react';

import { Button } from '@/components/ui/button';
import { HoverCard, HoverCardContent, HoverCardTrigger } from '@/components/ui/hover-card';

const meta: Meta<typeof HoverCard> = {
  title: 'Containers/Hover Card',
  component: HoverCard,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  subcomponents: {
    HoverCardContent: HoverCardContent as any,
    HoverCardTrigger: HoverCardTrigger as any,
  },
  render: (props) => (
    <HoverCard {...props}>
      <HoverCardTrigger>Hover</HoverCardTrigger>
      <HoverCardContent>
        <p className='small'>Hello from the popover</p>
      </HoverCardContent>
    </HoverCard>
  ),
};

export default meta;
type Story = StoryObj<typeof HoverCard>;

export const Default: Story = {};
