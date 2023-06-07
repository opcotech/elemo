import { concat, formatErrorMessage } from '@/lib/helpers';

import { type FormCommonProps, FormFieldContainer } from './FormFieldContainer';
import type { ChangeEvent, ReactNode } from 'react';

export interface FormInputProps extends FormCommonProps {
  type: 'text' | 'email' | 'password' | 'number' | 'date' | 'time' | 'datetime-local' | 'tel' | 'url' | 'search';
  prefix?: ReactNode;
  addon?: ReactNode;
  addonPosition?: 'left' | 'right';
  addonClassName?: string;
  value?: string;
  onChange?: (event: ChangeEvent<HTMLInputElement>) => void;
}

export function FormInput(props: FormInputProps) {
  const error = props.errors[props.errorField ?? props.name];

  const registerProps = {
    ...props.register(props.name),
    onChange: props.onChange
  };

  return (
    <FormFieldContainer {...props}>
      <div className={concat(props.addon ? 'flex' : 'relative', 'rounded-md shadow-sm')}>
        {props.prefix && (
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
            <span className="text-gray-500 sm:text-sm">{props.prefix}</span>
          </div>
        )}
        {props.addon && props.addonPosition !== 'right' && (
          <span
            className={concat(
              'inline-flex items-center rounded-l-md border border-r-0 border-gray-300 px-3 text-gray-500 sm:text-sm',
              props.addonClassName
            )}
          >
            {props.addon}
          </span>
        )}
        <input
          type={props.type}
          className={concat(
            props.className,
            error
              ? 'text-red-800 border-red-300 focus:border-red-500 focus:ring-red-500'
              : 'border-gray-300 focus:border-gray-500 focus:ring-gray-500',
            props.prefix ? 'pl-8' : '',
            props.addon && props.addonPosition !== 'right' ? 'rounded-none rounded-r-md' : 'rounded-md',
            props.addon && props.addonPosition === 'right' ? 'rounded-none rounded-l-md' : 'rounded-md',
            props.disabled ? 'opacity-70 bg-gray-50 cursor-not-allowed' : '',
            'form-input block w-full sm:text-sm'
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
        {props.addon && props.addonPosition === 'right' && (
          <span
            className={concat(
              'inline-flex items-center rounded-r-md border border-l-0 border-gray-300 px-3 text-gray-500 sm:text-sm',
              props.addonClassName
            )}
          >
            {props.addon}
          </span>
        )}
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
