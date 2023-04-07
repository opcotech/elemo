'use client';

import {Transition} from '@headlessui/react';
import {Fragment, useState} from 'react';

import Icon from '@/components/Icon';
import {concat} from '@/helpers';
import useTimeout from '@/hooks/useTimeout';
import type {IconVariant} from '@/types/heroicon';

const variantIcons = {
  success: 'CheckCircleIcon',
  info: 'InformationCircleIcon',
  warning: 'ExclamationCircleIcon',
  error: 'XCircleIcon'
};

const variantClasses = {
  success: 'text-green-600',
  info: 'text-blue-600',
  warning: 'text-yellow-600',
  error: 'text-red-600'
};

export default function Message({type, title, message, dismissAfter}: Message) {
  const [show, setShow] = useState(true);
  useTimeout(() => setShow(false), dismissAfter && dismissAfter > 0 ? dismissAfter : 5000);

  return (
    <Transition
      show={show}
      as={Fragment}
      enter="transform ease-out duration-300 transition"
      enterFrom="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2"
      enterTo="translate-y-0 opacity-100 sm:translate-x-0"
      leave="transition ease-in-out duration-200"
      leaveFrom="opacity-100"
      leaveTo="opacity-0"
    >
      <div
        className="pointer-events-auto w-full max-w-sm overflow-hidden rounded-lg bg-white shadow-lg ring-1 ring-black ring-opacity-5">
        <div className="p-4">
          <div className="flex items-start">
            <div className="flex-shrink-0">
              <Icon
                variant={variantIcons[type] as IconVariant}
                className={concat(variantClasses[type], 'h-6 w-6')}
                aria-hidden="true"
              />
            </div>
            <div className="ml-3 w-0 flex-1 pt-0.5">
              <p className="text-sm font-medium text-gray-900">{title}</p>
              <p className="mt-1 text-sm text-gray-500">{message}</p>
            </div>
            {dismissAfter !== 0 && (
              <div className="ml-4 flex flex-shrink-0">
                <button
                  id={`btn-modal-close`}
                  type="button"
                  className="rounded-full bg-white text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
                  onClick={() => setShow(false)}
                >
                  <span className="sr-only">Close panel</span>
                  <Icon variant="XMarkIcon" className="h-5 w-5" aria-hidden="true"/>
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </Transition>
  );
}
