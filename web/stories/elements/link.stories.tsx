import type { Meta, StoryObj } from '@storybook/react';

import { Link } from '@/components/ui/link';

const meta: Meta<typeof Link> = {
  title: 'Elements/Link',
  component: Link,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  render: (props) => {
    return <Link {...props}>create beautiful links</Link>;
  },
};

export default meta;
type Story = StoryObj<typeof Link>;

export const Default: Story = {};

export const Destructive: Story = {
  args: {
    variant: 'destructive',
  },
};

export const External: Story = {
  args: {
    isExternal: true,
  },
};
