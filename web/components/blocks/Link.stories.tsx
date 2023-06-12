import type { Meta, StoryObj } from '@storybook/react';

import { Link } from './Link';

const meta: Meta<typeof Link> = {
  title: 'Elements/Link',
  component: Link,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof Link>;

export const Sample: Story = {
  args: {
    href: 'https://example.com',
    children: 'https://example.com'
  }
};

export const Decorated: Story = {
  args: {
    href: 'https://example.com',
    decorated: true,
    children: 'https://example.com'
  }
};

export const NotDecorated: Story = {
  args: {
    href: 'https://example.com',
    decorated: false,
    children: 'https://example.com'
  }
};
