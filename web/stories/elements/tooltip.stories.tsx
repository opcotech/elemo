import type { Meta, StoryObj } from '@storybook/react';

import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

const meta: Meta<typeof TooltipContent> = {
  title: 'Elements/Tooltip',
  component: TooltipContent,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  render: (props) => {
    return (
      <TooltipProvider>
        <Tooltip defaultOpen={true}>
          <TooltipTrigger>
            <p>Hover over me</p>
          </TooltipTrigger>
          <TooltipContent {...props}>Hello!</TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  },
};

export default meta;
type Story = StoryObj<typeof TooltipContent>;

export const Default: Story = {};

export const Red: Story = {
  args: {
    variant: 'red',
  },
};

export const Yellow: Story = {
  args: {
    variant: 'yellow',
  },
};

export const Green: Story = {
  args: {
    variant: 'green',
  },
};

export const Blue: Story = {
  args: {
    variant: 'blue',
  },
};
