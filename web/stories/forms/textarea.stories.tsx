import type { Meta, StoryObj } from '@storybook/react';

import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';

const meta: Meta<typeof Textarea> = {
  title: 'Forms/Textarea',
  component: Textarea,
  tags: ['autodocs'],
  args: {
    placeholder: 'type something...',
  },
  render: (props) => {
    return (
      <div className='mx-auto grid w-full max-w-lg items-center gap-1.5'>
        <Textarea id='input-field' {...props} />
      </div>
    );
  },
};

export default meta;
type Story = StoryObj<typeof Textarea>;

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
        <Textarea id='input-field' {...props} />
      </div>
    );
  },
};
