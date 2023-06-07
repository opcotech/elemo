import type { FieldErrorsImpl, UseFormRegister } from 'react-hook-form';

import { concat } from '@/lib/helpers';
import type { KeyboardEvent, ReactNode } from 'react';

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

export function FormFieldContainer({ grid = true, name, label, required, children }: FormCommonProps) {
  return (
    <div className={concat(grid ? 'sm:grid sm:grid-cols-12 sm:items-start sm:gap-3' : '')}>
      <label
        htmlFor={name}
        className={concat(grid ? '' : 'mb-2', 'block text-sm font-medium text-gray-700 sm:mt-px sm:pt-2 sm:col-span-3')}
      >
        {label}
        {required && <span className="text-xs text-red-500 ml-0.5">*</span>}
      </label>
      <div className="sm:col-span-9 sm:mt-0">{children}</div>
    </div>
  );
}
