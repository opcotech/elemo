import type { Meta, StoryObj } from '@storybook/react';

import { FormTextarea, FormTextareaProps } from './FormTextarea';
import { useForm } from 'react-hook-form';

const meta: Meta<typeof FormTextarea> = {
  title: 'Elements/FormTextarea',
  component: FormTextarea,
  tags: ['autodocs'],
  args: {
    label: 'Label',
    name: 'textarea',
    placeholder: 'placeholder',
    rows: 5,
    grid: false,
    required: false,
    disabled: false,
    errors: {}
  }
};

export default meta;
type Story = StoryObj<typeof FormTextarea>;

export const Sample = (args: Story['args']) => {
  const { register } = useForm();
  return (
    <div>
      <FormTextarea {...(args as FormTextareaProps)} register={register} />
    </div>
  );
};

export const RowLayout = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'textarea',
    placeholder: 'placeholder',
    rows: 5,
    grid: false,
    required: false,
    disabled: false,
    errors: {}
  };

  return (
    <div>
      <FormTextarea {...(args as FormTextareaProps)} register={register} />
    </div>
  );
};

export const GridLayout = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'textarea',
    placeholder: 'placeholder',
    rows: 5,
    grid: true,
    required: false,
    disabled: false,
    errors: {}
  };

  return (
    <div>
      <FormTextarea {...(args as FormTextareaProps)} register={register} />
    </div>
  );
};

export const Required = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'textarea',
    placeholder: 'placeholder',
    rows: 5,
    grid: false,
    required: true,
    disabled: false,
    errors: {}
  };

  return (
    <div>
      <FormTextarea {...(args as FormTextareaProps)} register={register} />
    </div>
  );
};

export const Disabled = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'textarea',
    placeholder: 'placeholder',
    rows: 5,
    grid: false,
    required: false,
    disabled: true,
    errors: {}
  };

  return (
    <div>
      <FormTextarea {...(args as FormTextareaProps)} register={register} />
    </div>
  );
};

export const WithErrors = () => {
  const { register } = useForm();
  const args = {
    label: 'Label',
    name: 'textarea',
    placeholder: 'placeholder',
    rows: 5,
    grid: false,
    required: true,
    disabled: false
  };

  return (
    <div>
      <FormTextarea
        {...(args as FormTextareaProps)}
        register={register}
        errors={{
          textarea: {
            type: 'required',
            message: 'This field is required'
          }
        }}
      />
    </div>
  );
};
