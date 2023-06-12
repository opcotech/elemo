import { Switch } from '@headlessui/react';

import { concat } from '@/lib/helpers';

export interface FormSwitchProps {
  label: string;
  name?: string;
  description?: string;
  checked?: boolean;
  disabled?: boolean;
  grid?: boolean;
  onChange?: (checked: boolean) => void;
}

export function FormSwitch({ label, name, description, checked, disabled, grid = true, onChange }: FormSwitchProps) {
  return (
    <div className={concat(grid ? 'sm:grid sm:grid-cols-12 sm:items-start sm:gap-3' : '')}>
      <div className={concat(grid ? 'mt-1 sm:col-span-9 sm:col-start-4 sm:mt-0' : '')}>
        <Switch.Group
          as="div"
          className={concat('flex space-x-6 justify-between', description ? 'items-start' : 'items-center')}
        >
          <Switch
            name={name}
            onChange={onChange}
            disabled={disabled}
            className={concat(
              checked ? 'bg-gray-600' : 'bg-gray-200',
              disabled ? 'opacity-70 cursor-not-allowed' : '',
              'relative inline-flex mt-1 h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent   transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2'
            )}
          >
            <span
              aria-hidden="true"
              className={concat(
                checked ? 'translate-x-5' : 'translate-x-0',
                'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out'
              )}
            />
          </Switch>
          <span className="flex flex-grow flex-col">
            <Switch.Label
              as="span"
              className={concat('text-sm font-medium text-gray-900', description ? '' : 'pt-1')}
              passive
            >
              {label}
            </Switch.Label>
            {description && (
              <Switch.Description as="span" className="text-sm text-gray-500">
                {description}
              </Switch.Description>
            )}
          </span>
        </Switch.Group>
      </div>
    </div>
  );
}
