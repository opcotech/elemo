import { Combobox } from '@headlessui/react';
import type { ReactNode, SetStateAction } from 'react';
import { concat, formatErrorMessage } from '@/lib/helpers';
import { type FormCommonProps, FormFieldContainer } from './FormFieldContainer';
import { Icon } from '@/components/blocks/Icon';
import { Badge } from '@/components/blocks/Badge';

export interface FormSelectOption {
  label: string;
  value: any;
}

export interface FormSelectProps extends FormCommonProps {
  multiple?: boolean;
  options: FormSelectOption[];
  selectedOptions: FormSelectOption | FormSelectOption[] | undefined;
  setSelectedOptions: (value: SetStateAction<any>) => void;
  setFilter: (value: SetStateAction<string>) => void;
  placeholder?: string;
  required?: boolean;
  children?: ReactNode;
}

export function FormSelect(props: FormSelectProps) {
  const error = props.errors[props.name];

  const displaySelected = (value?: FormSelectOption): string => {
    return props.multiple ? '' : value?.label || '';
  };

  function dismissSelection(item: FormSelectOption) {
    if (!props.multiple) return;
    props.setSelectedOptions((props.selectedOptions as FormSelectOption[]).filter((i) => i.value !== item.value));
  }

  return (
    <FormFieldContainer {...props}>
      <Combobox
        as="div"
        defaultValue={props.selectedOptions}
        onChange={props.setSelectedOptions}
        /* @ts-ignore */
        multiple={props.multiple}
      >
        <div className="relative mb-4">
          <Combobox.Input
            className={concat(
              error
                ? 'text-red-800 border-red-300 focus:border-red-500 focus:ring-red-500'
                : 'border-gray-300 focus:border-gray-500 focus:ring-gray-500',
              props.disabled ? 'opacity-70 bg-gray-50 cursor-not-allowed' : '',
              'form-input w-full rounded-md border bg-white py-2 pl-3 pr-10 shadow-sm focus:outline-none focus:ring-1 sm:text-sm'
            )}
            aria-disabled={props.disabled}
            onChange={(event) => props.setFilter(event.target.value)}
            displayValue={displaySelected}
            placeholder={props.placeholder}
            autoComplete="off"
          />

          <Combobox.Button
            id="btn-personal-settings-select-language"
            className="absolute inset-y-0 right-0 flex items-center rounded-r-md px-2 focus:outline-none"
          >
            <Icon size={'sm'} variant="ChevronUpDownIcon" className="h-4 w-4 text-gray-400" aria-hidden="true" />
          </Combobox.Button>

          <Combobox.Options className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
            {props.options.length === 0 && (
              <div className="relative cursor-default select-none py-2 px-4 text-gray-700">No such option.</div>
            )}

            {props.options.map((item) => (
              <Combobox.Option
                key={item.value}
                value={item}
                disabled={props.disabled}
                className={({ active, disabled }) =>
                  concat(
                    'relative cursor-default select-none py-1.5 pl-8 pr-3',
                    disabled
                      ? 'text-gray-400'
                      : active
                      ? 'bg-gray-50 text-blue-500'
                      : 'text-gray-700 hover:text-blue-500 hover:bg-gray-50'
                  )
                }
              >
                {({ active, selected, disabled }) => {
                  return (
                    <>
                      <span className="block truncate">{item.label}</span>

                      {selected && (
                        <span
                          className={concat(
                            'absolute inset-y-0 left-0 flex items-center pl-1.5',
                            disabled ? 'text-gray-400' : active ? 'text-blue-500' : 'text-gray-600'
                          )}
                        >
                          <Icon size={'sm'} variant="CheckIcon" />
                        </span>
                      )}
                    </>
                  );
                }}
              </Combobox.Option>
            ))}
          </Combobox.Options>
        </div>
      </Combobox>

      {props.multiple && props.selectedOptions && (
        <div className="flex mt-2 space-x-2">
          {(props.selectedOptions as FormSelectOption[]).map((item) => (
            <Badge
              key={item.value}
              title={item.label}
              className={'mb-2'}
              dismissible
              onDismiss={() => dismissSelection(item)}
            />
          ))}
        </div>
      )}

      {error && (
        <p id={`${props.name}-error`} className="mt-2 text-sm text-red-600">
          {formatErrorMessage(props.name, error.message as string)}
        </p>
      )}
      {props.children}
    </FormFieldContainer>
  );
}
