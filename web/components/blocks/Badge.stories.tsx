import type { Meta, StoryObj } from '@storybook/react';

import { Badge } from './Badge';

const meta: Meta<typeof Badge> = {
  title: 'Elements/Badge',
  component: Badge,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof Badge>;

export const Sample: Story = {
  args: {
    title: 'Sample',
    size: 'md',
    variant: 'neutral',
    dismissible: false,
    onDismiss: () => alert('Dismissed!')
  }
};

export const Neutral: Story = {
  args: {
    title: 'Sample',
    size: 'md',
    variant: 'neutral'
  }
};

export const Info: Story = {
  args: {
    title: 'Sample',
    size: 'md',
    variant: 'info'
  }
};

export const Success: Story = {
  args: {
    title: 'Sample',
    size: 'md',
    variant: 'success'
  }
};

export const Warning: Story = {
  args: {
    title: 'Sample',
    size: 'md',
    variant: 'warning'
  }
};

export const Danger: Story = {
  args: {
    title: 'Sample',
    size: 'md',
    variant: 'danger'
  }
};

export const Dismissible: Story = {
  args: {
    title: 'Sample',
    size: 'md',
    variant: 'neutral',
    dismissible: true,
    onDismiss: () => alert('Dismissed!')
  }
};
