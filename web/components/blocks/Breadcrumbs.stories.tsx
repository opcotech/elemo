import type { Meta, StoryObj } from '@storybook/react';

import { Breadcrumbs } from './Breadcrumbs';

const meta: Meta<typeof Breadcrumbs> = {
  title: 'Elements/Breadcrumbs',
  component: Breadcrumbs,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof Breadcrumbs>;

export const Sample: Story = {
  args: {
    links: [
      { name: 'Namespaces', href: '/namespaces' },
      { name: 'Elemo', href: '/namespaces/elemo' },
      { name: 'Projects', href: '/namespaces/elemo/projects' },
      { name: 'Project 1', href: '/namespaces/elemo/projects/1', current: true }
    ]
  }
};
