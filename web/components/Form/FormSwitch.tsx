import {Switch} from '@headlessui/react';

import {concat} from '@/helpers';

export interface FormSwitchProps {
  label: string;
  description?: string;
  checked?: boolean;
  onChange?: (checked: boolean) => void;
}

export default function FormSwitch({label, description, checked = false, onChange}: FormSwitchProps) {
  return (
    <div className="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4">
      <div className="mt-1 sm:col-start-2 sm:col-span-2 sm:mt-0">
        <Switch.Group as="div" className="flex space-x-6 items-start justify-between">
          <Switch
            id="switch-remote-work"
            checked={checked}
            onChange={onChange}
            className={concat(
              checked ? 'bg-gray-600' : 'bg-gray-200',
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
            <Switch.Label as="span" className="text-sm font-medium text-gray-900" passive>
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
