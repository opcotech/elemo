import type { Meta, StoryObj } from '@storybook/react';

import { Navbar } from './Navbar';
import { SessionProvider } from 'next-auth/react';
import { Session } from 'next-auth';

const FAKE_SESSION: Session = {
  user: {
    id: '1',
    name: 'John Doe',
    email: 'john.doe@example.com',
    image: 'https://picsum.photos/id/433/100/100',
    accessToken: '',
    refreshToken: '',
    accessTokenExpiresIn: 3600
  },
  expires: new Date().toISOString()
};

const meta: Meta<typeof Navbar> = {
  title: 'Navigation/Navbar',
  component: Navbar,
  tags: ['autodocs'],
  render: (args) => (
    <SessionProvider session={FAKE_SESSION}>
      <Navbar {...args} />
    </SessionProvider>
  )
};

export default meta;
type Story = StoryObj<typeof Navbar>;

export const Sample: Story = {
  args: {
    hasTodos: false,
    hasNotifications: false,
    navigation: [
      { id: '1', href: '#', label: 'Dashboard', prefetch: false },
      { id: '1', href: '#', label: 'Users', prefetch: false },
      { id: '1', href: '#', label: 'Settings', prefetch: false }
    ],
    userNavigation: [
      { id: '1', href: '#', label: 'Profile', prefetch: false },
      { id: '1', href: '#', label: 'Settings', prefetch: false },
      { id: '1', href: '#', label: 'Sign out', prefetch: false }
    ]
  }
};

export const HasTodos: Story = {
  args: {
    hasTodos: true,
    hasNotifications: false,
    navigation: [
      { id: '1', href: '#', label: 'Dashboard', prefetch: false },
      { id: '1', href: '#', label: 'Users', prefetch: false },
      { id: '1', href: '#', label: 'Settings', prefetch: false }
    ],
    userNavigation: [
      { id: '1', href: '#', label: 'Profile', prefetch: false },
      { id: '1', href: '#', label: 'Settings', prefetch: false },
      { id: '1', href: '#', label: 'Sign out', prefetch: false }
    ]
  }
};

export const HasNotifications: Story = {
  args: {
    hasTodos: false,
    hasNotifications: true,
    navigation: [
      { id: '1', href: '#', label: 'Dashboard', prefetch: false },
      { id: '1', href: '#', label: 'Users', prefetch: false },
      { id: '1', href: '#', label: 'Settings', prefetch: false }
    ],
    userNavigation: [
      { id: '1', href: '#', label: 'Profile', prefetch: false },
      { id: '1', href: '#', label: 'Settings', prefetch: false },
      { id: '1', href: '#', label: 'Sign out', prefetch: false }
    ]
  }
};
