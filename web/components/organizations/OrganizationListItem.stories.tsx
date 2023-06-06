import type { Meta, StoryObj } from '@storybook/react';

import { OrganizationStatus } from '@/lib/api';
import { OrganizationListItem } from './OrganizationListItem';

const meta: Meta<typeof OrganizationListItem> = {
  title: 'Organizations/OrganizationListItem',
  component: OrganizationListItem,
  tags: ['autodocs'],
  render: (args) => (
    <ul>
      <OrganizationListItem {...args} />
    </ul>
  )
};

export default meta;
type Story = StoryObj<typeof OrganizationListItem>;

export const Sample: Story = {
  args: {
    organization: {
      id: '1',
      name: 'ACME Inc.',
      email: 'info@example.com',
      status: OrganizationStatus.ACTIVE,
      logo: 'https://picsum.photos/id/85/100/100',
      website: 'https://example.com',
      members: ['1', '2', '3'],
      teams: ['1', '2'],
      namespaces: ['1'],
      created_at: new Date().toISOString(),
      updated_at: null
    },
    canView: true,
    canEdit: true,
    canDelete: true
  }
};

export const Active: Story = {
  args: {
    organization: {
      id: '1',
      name: 'ACME Inc.',
      email: 'info@example.com',
      status: OrganizationStatus.ACTIVE,
      logo: 'https://picsum.photos/id/85/100/100',
      website: 'https://example.com',
      members: ['1', '2', '3'],
      teams: ['1', '2'],
      namespaces: ['1'],
      created_at: new Date().toISOString(),
      updated_at: null
    },
    canView: true,
    canEdit: true,
    canDelete: true
  }
};

export const Deleted: Story = {
  args: {
    organization: {
      id: '1',
      name: 'ACME Inc.',
      email: 'info@example.com',
      status: OrganizationStatus.DELETED,
      logo: 'https://picsum.photos/id/85/100/100',
      website: 'https://example.com',
      members: ['1', '2', '3'],
      teams: ['1', '2'],
      namespaces: ['1'],
      created_at: new Date().toISOString(),
      updated_at: null
    },
    canView: true,
    canEdit: true,
    canDelete: true
  }
};

export const UserCanView: Story = {
  args: {
    organization: {
      id: '1',
      name: 'ACME Inc.',
      email: 'info@example.com',
      status: OrganizationStatus.ACTIVE,
      logo: 'https://picsum.photos/id/85/100/100',
      website: 'https://example.com',
      members: ['1', '2', '3'],
      teams: ['1', '2'],
      namespaces: ['1'],
      created_at: new Date().toISOString(),
      updated_at: null
    },
    canView: true,
    canEdit: false,
    canDelete: false
  }
};

export const UserCanEdit: Story = {
  args: {
    organization: {
      id: '1',
      name: 'ACME Inc.',
      email: 'info@example.com',
      status: OrganizationStatus.ACTIVE,
      logo: 'https://picsum.photos/id/85/100/100',
      website: 'https://example.com',
      members: ['1', '2', '3'],
      teams: ['1', '2'],
      namespaces: ['1'],
      created_at: new Date().toISOString(),
      updated_at: null
    },
    canView: true,
    canEdit: true,
    canDelete: false
  }
};

export const UserCanDelete: Story = {
  args: {
    organization: {
      id: '1',
      name: 'ACME Inc.',
      email: 'info@example.com',
      status: OrganizationStatus.ACTIVE,
      logo: 'https://picsum.photos/id/85/100/100',
      website: 'https://example.com',
      members: ['1', '2', '3'],
      teams: ['1', '2'],
      namespaces: ['1'],
      created_at: new Date().toISOString(),
      updated_at: null
    },
    canView: true,
    canEdit: true,
    canDelete: true
  }
};
