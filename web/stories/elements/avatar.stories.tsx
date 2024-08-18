import type { Meta, StoryObj } from '@storybook/react';

import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar';

const meta: Meta<typeof Avatar> = {
  title: 'Elements/Avatar',
  component: Avatar,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  subcomponents: {
    AvatarImage: AvatarImage as any,
    AvatarFallback: AvatarFallback as any,
  },
};

export default meta;
type Story = StoryObj<typeof Avatar>;

export const Default: Story = {
  render: (props) => {
    return (
      <Avatar {...props}>
        <AvatarImage src='https://i.pravatar.cc/150?img=8' alt='Avatar of John Doe' />
        <AvatarFallback>JD</AvatarFallback>
      </Avatar>
    );
  },
};

export const WithoutImage: Story = {
  render: (props) => {
    return (
      <Avatar {...props}>
        <AvatarFallback>JD</AvatarFallback>
      </Avatar>
    );
  },
};
