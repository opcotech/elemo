import type { Meta, StoryObj } from '@storybook/react';

import { Switch } from '@/components/ui/switch';

const meta: Meta<typeof Switch> = {
  title: 'Forms/Switch',
  component: Switch,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  render: (props) => {
    return (
      <div className='items-top flex space-x-2'>
        <Switch id='terms1' {...props} />
        <div className='grid gap-1.5 leading-none'>
          <label
            htmlFor='terms1'
            className='text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70'
          >
            Accept terms and conditions
          </label>
          <p className='text-sm text-muted-foreground'>You agree to our Terms of Service and Privacy Policy.</p>
        </div>
      </div>
    );
  },
};

export default meta;
type Story = StoryObj<typeof Switch>;

export const Default: Story = {};

export const Disabled: Story = {
  args: {
    checked: true,
    disabled: true,
  },
};
