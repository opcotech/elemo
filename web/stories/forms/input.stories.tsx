import type { Meta, StoryObj } from '@storybook/react';

import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

const meta: Meta<typeof Input> = {
  title: 'Forms/Input',
  component: Input,
  tags: ['autodocs'],
  args: {
    type: 'text',
    placeholder: 'type something...',
  },
  render: (props) => {
    return (
      <div className='mx-auto grid w-full max-w-lg items-center gap-1.5'>
        <Input id='input-field' {...props} />
      </div>
    );
  },
};

export default meta;
type Story = StoryObj<typeof Input>;

export const Default: Story = {};

export const Disabled: Story = {
  args: {
    disabled: true,
  },
};

export const WitLabel: Story = {
  render: (props) => {
    return (
      <div className='mx-auto grid w-full max-w-lg items-center gap-1.5'>
        <Label htmlFor='input-field'>Input field</Label>
        <Input id='input-field' {...props} />
      </div>
    );
  },
};
