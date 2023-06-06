import type { Meta, StoryObj } from '@storybook/react';

import { FormSwitch, FormSwitchProps } from './FormSwitch';
import { useForm } from 'react-hook-form';
import { useState } from 'react';

const meta: Meta<typeof FormSwitch> = {
  title: 'Elements/FormSwitch',
  component: FormSwitch,
  tags: ['autodocs'],
  args: {
    label: 'Label',
    description: 'A description to help the user.',
    disabled: false
  }
};

export default meta;
type Story = StoryObj<typeof FormSwitch>;

export const Sample = (args: Story['args']) => {
  const [checked, setChecked] = useState(false);

  return (
    <div>
      <FormSwitch {...(args as FormSwitchProps)} checked={checked} onChange={(val: boolean) => setChecked(val)} />
    </div>
  );
};

export const Checked = (args: Story['args']) => {
  const [checked, setChecked] = useState(true);

  return (
    <div>
      <FormSwitch {...(args as FormSwitchProps)} checked={checked} onChange={(val: boolean) => setChecked(val)} />
    </div>
  );
};

export const Disabled = (args: Story['args']) => {
  return (
    <div>
      <FormSwitch {...(args as FormSwitchProps)} disabled={true} />
    </div>
  );
};

export const NoDescription = (args: Story['args']) => {
  const [checked, setChecked] = useState(false);

  return (
    <div>
      <FormSwitch label={'Label'} checked={checked} onChange={(val: boolean) => setChecked(val)} />
    </div>
  );
};
