import type { Meta, StoryObj } from '@storybook/react';

import { ImageSkeleton } from './ImageSkeleton';

const meta: Meta<typeof ImageSkeleton> = {
  title: 'Elements/ImageSkeleton',
  component: ImageSkeleton,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof ImageSkeleton>;

export const Sample: Story = {
  args: {}
};

export const CustomClasses: Story = {
  args: {
    className: 'w-32 h-32'
  }
};
