import type { Meta, StoryObj } from '@storybook/react';

import { LineSkeleton } from './LineSkeleton';

const meta: Meta<typeof LineSkeleton> = {
  title: 'Elements/LineSkeleton',
  component: LineSkeleton,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof LineSkeleton>;

export const Sample: Story = {
  args: {}
};
