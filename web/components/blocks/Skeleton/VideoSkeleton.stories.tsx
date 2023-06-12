import type { Meta, StoryObj } from '@storybook/react';

import { VideoSkeleton } from './VideoSkeleton';

const meta: Meta<typeof VideoSkeleton> = {
  title: 'Elements/VideoSkeleton',
  component: VideoSkeleton,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof VideoSkeleton>;

export const Sample: Story = {
  args: {}
};

export const CustomClasses: Story = {
  args: {
    className: 'w-32 h-32'
  }
};
