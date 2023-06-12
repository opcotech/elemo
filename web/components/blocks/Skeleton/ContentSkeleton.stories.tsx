import type { Meta, StoryObj } from '@storybook/react';

import { ContentSkeleton } from './ContentSkeleton';

const meta: Meta<typeof ContentSkeleton> = {
  title: 'Elements/ContentSkeleton',
  component: ContentSkeleton,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof ContentSkeleton>;

export const Sample: Story = {
  args: {}
};
