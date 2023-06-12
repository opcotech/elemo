import type { Meta, StoryObj } from '@storybook/react';

import { FormInput, type FormInputProps } from './FormInput';
import { useForm } from 'react-hook-form';
import { Icon } from '@/components/blocks/Icon';
import { Button } from '@/components/blocks/Button';

const meta: Meta<typeof FormInput> = {
  title: 'Elements/FormInput',
  component: FormInput,
  tags: ['autodocs'],
  args: {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: false,
    disabled: false,
    errors: {}
  }
};

export default meta;
type Story = StoryObj<typeof FormInput>;

export const Sample = (args: Story['args']) => {
  const { register } = useForm();
  return (
    <div>
      <FormInput {...(args as FormInputProps)} register={register} />
    </div>
  );
};

export const RowLayout = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: false,
    disabled: false,
    errors: {}
  };

  return (
    <div>
      <FormInput {...(args as FormInputProps)} register={register} />
    </div>
  );
};

export const GridLayout = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: true,
    required: false,
    disabled: false,
    errors: {}
  };

  return (
    <div>
      <FormInput {...(args as FormInputProps)} register={register} />
    </div>
  );
};

export const Required = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: true,
    disabled: false,
    errors: {}
  };

  return (
    <div>
      <FormInput {...(args as FormInputProps)} register={register} />
    </div>
  );
};

export const Disabled = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: false,
    disabled: true,
    errors: {}
  };

  return (
    <div>
      <FormInput {...(args as FormInputProps)} register={register} />
    </div>
  );
};

export const WithErrors = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: true,
    disabled: false
  };

  return (
    <div>
      <FormInput
        {...(args as FormInputProps)}
        register={register}
        errors={{
          field: {
            type: 'required',
            message: 'This field is required'
          }
        }}
      />
    </div>
  );
};

export const WithPrefix = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: false,
    disabled: false,
    prefix: <Icon size={'xs'} variant={'RocketLaunchIcon'} />,
    errors: {}
  };

  return (
    <div>
      <FormInput {...(args as FormInputProps)} register={register} />
    </div>
  );
};

export const WithAddonOnLeft = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: false,
    disabled: false,
    addon: 'https://',
    errors: {}
  };

  return (
    <div>
      <FormInput {...(args as FormInputProps)} register={register} />
    </div>
  );
};

export const WithAddonOnRight = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: false,
    disabled: false,
    addon: (
      <Button size={'xs'} variant={'link'}>
        add
      </Button>
    ),
    addonPosition: 'right',
    addonClassName: 'hover:bg-gray-50',
    errors: {}
  };

  return (
    <div>
      <FormInput {...(args as FormInputProps)} register={register} />
    </div>
  );
};
