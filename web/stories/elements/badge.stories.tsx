import type { Meta, StoryObj } from '@storybook/react';

import { Badge } from '@/components/ui/badge';

const meta: Meta<typeof Badge> = {
  title: 'Elements/Badge',
  component: Badge,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  render: (props) => {
    return <Badge {...props}>Button</Badge>;
  },
};

export default meta;
type Story = StoryObj<typeof Badge>;

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

export const Purple: Story = {
  args: {
    variant: 'purple',
  },
};

export const Pink: Story = {
  args: {
    variant: 'pink',
  },
};
