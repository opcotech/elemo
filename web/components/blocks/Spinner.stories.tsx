import type { Meta, StoryObj } from '@storybook/react';

import { Spinner } from './Spinner';

const meta: Meta<typeof Spinner> = {
  title: 'Elements/Spinner',
  component: Spinner,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof Spinner>;

export const Sample: Story = {
  args: {
    className: 'h-10 w-10'
  }
};

export const Colored: Story = {
  args: {
    className: 'h-10 w-10 text-blue-500'
  }
};
