import type { Meta, StoryObj } from '@storybook/react';
import {
  IconFile,
  IconLayoutKanban,
  IconStack2,
  IconMap,
  IconComponents,
  IconRocket,
  IconSettings,
} from '@tabler/icons-react';

import { Layout } from '@/components/layouts/with-aside';

const meta: Meta<typeof Layout> = {
  title: 'Layouts/Default with aside',
  component: Layout,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
  args: {
    asideItems: [
      { icon: <IconStack2 stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Backlog' },
      { icon: <IconLayoutKanban stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Boards' },
      { icon: <IconMap stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Roadmap' },
      { icon: <IconRocket stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Releases' },
      { icon: <IconFile stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Documents' },
      { icon: <IconComponents stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Components' },
      { icon: <IconSettings stroke={1.5} className='h-5 w-5' />, href: '#', label: 'Settings' },
    ],
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
