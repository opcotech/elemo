import type {FieldErrorsImpl, UseFormRegister} from 'react-hook-form';

import {concat} from '@/helpers';
import type {KeyboardEvent, ReactNode} from 'react';

export interface FormCommonProps {
  label: string;
  name: string;
  grid?: boolean;
  placeholder?: string;
  required?: boolean;
  disabled?: boolean;
  register: UseFormRegister<any>;
  errors: Partial<FieldErrorsImpl>;
  children?: ReactNode;
  className?: string;
  onKeyDown?: (event: KeyboardEvent<HTMLInputElement>) => void;
}

export interface FormFieldProps {
  label: string;
  name: string;
  grid?: boolean;
  children?: ReactNode;
}

export default function FormFieldContainer({ grid = true, name, label, children }: FormFieldProps) {
  return (
    <div className={concat(grid ? 'sm:grid sm:grid-cols-3 sm:items-start sm:gap-4' : '')}>
      <label
        htmlFor={name}
        className={concat(grid ? '' : 'mb-2', 'block text-sm font-medium text-gray-700 sm:mt-px sm:pt-2')}
      >
        {label}
      </label>
      <div className="sm:col-span-2 sm:mt-0">{children}</div>
    </div>
  );
}
