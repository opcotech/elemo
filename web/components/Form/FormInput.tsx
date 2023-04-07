import {concat, formatErrorMessage} from '@/helpers';

import FormFieldContainer, {type FormCommonProps} from './FormFieldContainer';
import type {ChangeEvent, ReactNode} from "react";

export interface FormInputProps extends FormCommonProps {
  type: 'text' | 'email' | 'password' | 'number' | 'date' | 'time' | 'datetime-local' | 'tel' | 'url' | 'search';
  prefix?: ReactNode;
  value?: string;
  onChange?: (event: ChangeEvent<HTMLInputElement>) => void;
}

export default function FormInput(props: FormInputProps) {
  const error = props.errors[props.name];

  const registerProps = {
    ...props.register(props.name),
    onChange: props.onChange
  };

  return (
    <FormFieldContainer name={props.name} label={props.label} grid={props.grid}>
      <div className="relative rounded-md shadow-sm">
        {props.prefix && (
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
            <span className="text-gray-500 sm:text-sm">{props.prefix}</span>
          </div>
        )}
        <input
          type={props.type}
          className={concat(
            props.className,
            error
              ? 'text-red-800 border-red-300 focus:border-red-500 focus:ring-red-500'
              : 'border-gray-300 focus:border-gray-500 focus:ring-gray-500',
            props.prefix ? 'pl-8' : '',
            'block w-full rounded-md sm:text-sm'
          )}
          value={props.value}
          placeholder={props.placeholder}
          required={props.required}
          disabled={props.disabled}
          aria-invalid={error ? 'true' : 'false'}
          aria-describedby={error ? `${props.name}-error` : undefined}
          onKeyDown={props.onKeyDown}
          {...registerProps}
        />
      </div>
      {error && (
        <p id={`${props.name}-error`} className="mt-2 text-sm text-red-600">
          {formatErrorMessage(props.name, error.message as string)}
        </p>
      )}
      {props.children}
    </FormFieldContainer>
  );
}
