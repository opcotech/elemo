import type { Meta, StoryObj } from '@storybook/react';

import { FormSelect, FormSelectOption, FormSelectProps } from './FormSelect';
import { useForm } from 'react-hook-form';
import { useState } from 'react';
import { FormTextarea, FormTextareaProps } from '@/components/blocks/Form/FormTextarea';

const meta: Meta<typeof FormSelect> = {
  title: 'Elements/FormSelect',
  component: FormSelect,
  tags: ['autodocs'],
  args: {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: false,
    disabled: false,
    errors: {},
    multiple: false
  }
};

const items: FormSelectOption[] = [
  { label: 'Item 1', value: 'item1' },
  { label: 'Item 2', value: 'item2' },
  { label: 'Item 3', value: 'item3' }
];

export default meta;
type Story = StoryObj<typeof FormSelect>;

export const Sample = (args: Story['args']) => {
  const { register } = useForm();
  const [selected, setSelected] = useState<FormSelectOption | FormSelectOption[] | undefined>(undefined);
  const [filter, setFilter] = useState<string>('');

  const filteredItems = items.filter((item) => item.label.toLowerCase().includes(filter.toLowerCase())) || items;

  return (
    <div>
      <FormSelect
        {...(args as FormSelectProps)}
        options={filteredItems}
        selectedOptions={selected}
        setSelectedOptions={setSelected}
        setFilter={setFilter}
        register={register}
      />
    </div>
  );
};

export const RowLayout = () => {
  const { register } = useForm();
  const [selected, setSelected] = useState<FormSelectOption | FormSelectOption[] | undefined>(undefined);
  const [filter, setFilter] = useState<string>('');

  const filteredItems = items.filter((item) => item.label.toLowerCase().includes(filter.toLowerCase())) || items;

  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: false,
    disabled: false,
    errors: {},
    multiple: false
  };

  return (
    <div>
      <FormSelect
        {...(args as FormSelectProps)}
        options={filteredItems}
        selectedOptions={selected}
        setSelectedOptions={setSelected}
        setFilter={setFilter}
        register={register}
      />
    </div>
  );
};

export const GridLayout = () => {
  const { register } = useForm();
  const [selected, setSelected] = useState<FormSelectOption | FormSelectOption[] | undefined>(undefined);
  const [filter, setFilter] = useState<string>('');

  const filteredItems = items.filter((item) => item.label.toLowerCase().includes(filter.toLowerCase())) || items;

  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: true,
    required: false,
    disabled: false,
    errors: {},
    multiple: false
  };

  return (
    <div>
      <FormSelect
        {...(args as FormSelectProps)}
        options={filteredItems}
        selectedOptions={selected}
        setSelectedOptions={setSelected}
        setFilter={setFilter}
        register={register}
      />
    </div>
  );
};

export const Required = () => {
  const { register } = useForm();
  const [selected, setSelected] = useState<FormSelectOption | FormSelectOption[] | undefined>(undefined);
  const [filter, setFilter] = useState<string>('');

  const filteredItems = items.filter((item) => item.label.toLowerCase().includes(filter.toLowerCase())) || items;

  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: true,
    disabled: false,
    errors: {},
    multiple: false
  };

  return (
    <div>
      <FormSelect
        {...(args as FormSelectProps)}
        options={filteredItems}
        selectedOptions={selected}
        setSelectedOptions={setSelected}
        setFilter={setFilter}
        register={register}
      />
    </div>
  );
};

export const Disabled = () => {
  const { register } = useForm();
  const [selected, setSelected] = useState<FormSelectOption | FormSelectOption[] | undefined>(undefined);
  const [filter, setFilter] = useState<string>('');

  const filteredItems = items.filter((item) => item.label.toLowerCase().includes(filter.toLowerCase())) || items;

  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: false,
    disabled: true,
    errors: {},
    multiple: false
  };

  return (
    <div>
      <FormSelect
        {...(args as FormSelectProps)}
        options={filteredItems}
        selectedOptions={selected}
        setSelectedOptions={setSelected}
        setFilter={setFilter}
        register={register}
      />
    </div>
  );
};

export const WithErrors = () => {
  const { register } = useForm();
  const [selected, setSelected] = useState<FormSelectOption | FormSelectOption[] | undefined>(undefined);
  const [filter, setFilter] = useState<string>('');

  const filteredItems = items.filter((item) => item.label.toLowerCase().includes(filter.toLowerCase())) || items;

  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: false,
    disabled: false,
    multiple: false
  };

  return (
    <div>
      <FormSelect
        {...(args as FormSelectProps)}
        options={filteredItems}
        selectedOptions={selected}
        setSelectedOptions={setSelected}
        setFilter={setFilter}
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

export const MultiSelect = () => {
  const { register } = useForm();
  const [selected, setSelected] = useState<FormSelectOption | FormSelectOption[] | undefined>(undefined);
  const [filter, setFilter] = useState<string>('');

  const filteredItems = items.filter((item) => item.label.toLowerCase().includes(filter.toLowerCase())) || items;

  const args = {
    label: 'Label',
    name: 'field',
    placeholder: 'placeholder',
    grid: false,
    required: false,
    disabled: false,
    errors: {},
    multiple: true
  };

  return (
    <div>
      <FormSelect
        {...(args as FormSelectProps)}
        options={filteredItems}
        selectedOptions={selected}
        setSelectedOptions={setSelected}
        setFilter={setFilter}
        register={register}
      />
    </div>
  );
};
