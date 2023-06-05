import type { Meta, StoryObj } from '@storybook/react';

import { Avatar } from './Avatar';

const meta: Meta<typeof Avatar> = {
  title: 'Elements/Avatar',
  component: Avatar,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof Avatar>;

export const Sample: Story = {
  args: {
    size: 'md',
    src: 'https://picsum.photos/id/433/100/100',
    initials: 'JD',
    alt: 'John Doe Avatar'
  }
};

export const WithImage: Story = {
  args: {
    size: 'md',
    src: 'https://picsum.photos/id/433/100/100',
    initials: 'JD',
    alt: 'John Doe Avatar'
  }
};

export const NoImage: Story = {
  args: {
    size: 'md',
    initials: 'JD',
    alt: 'John Doe Avatar'
  }
};

export const NoData: Story = {
  args: {
    size: 'md'
  }
};
