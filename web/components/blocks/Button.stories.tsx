import type { Meta, StoryObj } from '@storybook/react';

import { Button } from './Button';

const meta: Meta<typeof Button> = {
  title: 'Elements/Button',
  component: Button,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof Button>;

export const Sample: Story = {
  args: {
    type: 'button',
    size: 'md',
    variant: 'primary',
    disabled: false,
    loading: false,
    children: 'Button'
  }
};

export const Disabled: Story = {
  args: {
    type: 'button',
    size: 'md',
    variant: 'primary',
    disabled: true,
    loading: false,
    children: 'Button'
  }
};

export const Loading: Story = {
  args: {
    type: 'button',
    size: 'md',
    variant: 'primary',
    disabled: false,
    loading: true,
    children: 'Button'
  }
};

export const IconButton: Story = {
  args: {
    type: 'button',
    size: 'sm',
    disabled: false,
    loading: false,
    icon: 'RocketLaunchIcon',
    'aria-label': 'Rocket Launch'
  }
};
