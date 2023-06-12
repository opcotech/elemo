import type { Meta, StoryObj } from '@storybook/react';

import { ListSkeleton } from './ListSkeleton';

const meta: Meta<typeof ListSkeleton> = {
  title: 'Elements/ListSkeleton',
  component: ListSkeleton,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof ListSkeleton>;

export const Sample: Story = {
  args: {
    withBorder: false,
    fullWidth: false,
    count: 3
  }
};

export const WithBorder: Story = {
  args: {
    withBorder: true,
    fullWidth: false,
    count: 3
  }
};

export const FullWidth: Story = {
  args: {
    withBorder: false,
    fullWidth: true,
    count: 3
  }
};
