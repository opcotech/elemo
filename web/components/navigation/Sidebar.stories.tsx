import type { Meta, StoryObj } from '@storybook/react';

import { Sidebar } from './Sidebar';

const meta: Meta<typeof Sidebar> = {
  title: 'Navigation/Sidebar',
  component: Sidebar,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof Sidebar>;

export const Sample: Story = {
  args: {
    navigation: [
      { href: '#', label: 'Dashboard', icon: 'HomeIcon' },
      { href: '#', label: 'Users', icon: 'UserGroupIcon' },
      { href: '#', label: 'Settings', icon: 'CogIcon' }
    ]
  }
};

export const NoIcons: Story = {
  args: {
    navigation: [
      { href: '#', label: 'Dashboard' },
      { href: '#', label: 'Users' },
      { href: '#', label: 'Settings' }
    ]
  }
};
