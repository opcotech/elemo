import type { Meta, StoryObj } from '@storybook/react';

import { Layout } from '@/components/layouts/default';

const meta: Meta<typeof Layout> = {
  title: 'Layouts/Default',
  component: Layout,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
  render: (props) => (
    <Layout {...props}>
      <h1>Lorem ipsum dolor sit amet</h1>
      <p>
        Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore
        magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo
        consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla
        pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est
        laborum.
      </p>
    </Layout>
  ),
};

export default meta;
type Story = StoryObj<typeof Layout>;

export const Default: Story = {};
