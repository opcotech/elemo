import type { Meta, StoryObj } from '@storybook/react';

import { TextSkeleton } from './TextSkeleton';

const meta: Meta<typeof TextSkeleton> = {
  title: 'Elements/TextSkeleton',
  component: TextSkeleton,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof TextSkeleton>;

export const Sample: Story = {
  args: {}
};
