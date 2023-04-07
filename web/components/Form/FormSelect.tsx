import {Combobox} from '@headlessui/react';
import type {ReactNode, SetStateAction} from 'react';
import type {FieldErrorsImpl} from 'react-hook-form/dist/types';

import Icon from '@/components/Icon';
import {concat, formatErrorMessage} from '@/helpers';

import FormFieldContainer, {type FormCommonProps} from './FormFieldContainer';

export interface FormSelectItem {
  label: string;
  value: any;
}

export interface FormSelectProps extends FormCommonProps {
  filteredItems: FormSelectItem[] | undefined;
  selected: FormSelectItem[] | FormSelectItem | undefined;
  placeholder?: string;
  required?: boolean;
  multiple?: boolean;
  errors: Partial<FieldErrorsImpl>;
  children?: ReactNode;
  handleFilter: (value: SetStateAction<string>) => void;
  handleSelect: (value: never) => void;
}

export default function FormSelect(props: FormSelectProps) {
  const error = props.errors[props.name];

  return (
    <FormFieldContainer name={props.name} label={props.label} grid={props.grid}>
      {/* @ts-ignore */}
      <Combobox as="div" value={props.selected} onChange={props.handleSelect} multiple={props.multiple}>
        <div className="relative mb-4">
          <Combobox.Input
            className="w-full rounded-md border border-gray-300 bg-white py-2 pl-3 pr-10 shadow-sm focus:border-gray-500 focus:outline-none focus:ring-1 focus:ring-gray-500 sm:text-sm"
            onChange={(event) => props.handleFilter(event.target.value)}
            placeholder={props.placeholder}
            autoComplete="off"
          />

          <Combobox.Button
            id="btn-personal-settings-select-language"
            className="absolute inset-y-0 right-0 flex items-center rounded-r-md px-2 focus:outline-none"
          >
            <Icon variant="ChevronUpDownIcon" className="h-4 w-4 text-gray-400" aria-hidden="true"/>
          </Combobox.Button>

          <Combobox.Options
            className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
            {props.filteredItems &&
              props.filteredItems.map((item) => (
                <Combobox.Option
                  key={item.value}
                  value={item}
                  className={({active}) =>
                    concat(
                      'relative cursor-default select-none py-2 pl-8 pr-4',
                      active ? 'bg-gray-600 text-white' : 'text-gray-900'
                    )
                  }
                >
                  {({active, selected}) => (
                    <>
                      <span className={concat('block truncate', selected ? 'font-medium' : '')}>{item.label}</span>

                      {selected && (
                        <span
                          className={concat(
                            'absolute inset-y-0 left-0 flex items-center pl-1.5',
                            active ? 'text-white' : 'text-gray-600'
                          )}
                        >
                          <Icon variant="CheckIcon" className="h-5 w-5" aria-hidden="true"/>
                        </span>
                      )}
                    </>
                  )}
                </Combobox.Option>
              ))}
          </Combobox.Options>
        </div>
      </Combobox>
      {error && (
        <p id={`${props.name}-error`} className="mt-2 text-sm text-red-600">
          {formatErrorMessage(props.name, error.message as string)}
        </p>
      )}
      {props.children}
    </FormFieldContainer>
  );
}
