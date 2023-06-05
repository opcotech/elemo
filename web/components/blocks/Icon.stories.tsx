import type { Meta, StoryObj } from '@storybook/react';

import { Icon } from './Icon';

const meta: Meta<typeof Icon> = {
  title: 'Elements/Icon',
  component: Icon,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof Icon>;

export const Sample: Story = {
  args: {
    size: 'md',
    variant: 'RocketLaunchIcon'
  }
};
