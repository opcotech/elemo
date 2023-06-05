import type { Meta, StoryObj } from '@storybook/react';

import { Divider } from './Divider';

const meta: Meta<typeof Divider> = {
  title: 'Elements/Divider',
  component: Divider,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof Divider>;

export const Sample: Story = {
  args: {}
};
