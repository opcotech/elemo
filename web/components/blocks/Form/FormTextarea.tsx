import { concat, formatErrorMessage } from '@/lib/helpers';
import { type FormCommonProps, FormFieldContainer } from './FormFieldContainer';

export interface FormTextareaProps extends FormCommonProps {
  rows?: number;
}

export function FormTextarea(props: FormTextareaProps) {
  const error = props.errors[props.name];

  return (
    <FormFieldContainer {...props}>
      <textarea
        id={props.name}
        className={concat(
          props.className,
          error
            ? 'text-red-800 border-red-300 focus:border-red-500 focus:ring-red-500'
            : 'border-gray-300 focus:border-gray-500 focus:ring-gray-500',
          props.disabled ? 'opacity-70 bg-gray-50 cursor-not-allowed' : '',
          'block w-full rounded-md sm:text-sm'
        )}
        style={{ minHeight: '8rem' }}
        rows={props.rows}
        placeholder={props.placeholder}
        required={props.required}
        disabled={props.disabled}
        aria-invalid={error ? 'true' : 'false'}
        aria-describedby={error ? `${props.name}-error` : undefined}
        {...props.register(props.name)}
      />
      {error && (
        <p id={`${props.name}-error`} className="mt-2 text-sm text-red-600">
          {formatErrorMessage(props.name, error.message as string)}
        </p>
      )}
    </FormFieldContainer>
  );
}
