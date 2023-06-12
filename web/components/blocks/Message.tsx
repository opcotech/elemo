'use client';

import { Fragment, useState } from 'react';
import { Transition } from '@headlessui/react';
import type { IconVariant } from '@/types/heroicon';

import { concat } from '@/lib/helpers';
import useTimeout from '@/lib/hooks/useTimeout';
import { Icon } from '@/components/blocks/Icon';
import { Button } from './Button';

type MessageVariant = 'info' | 'success' | 'warning' | 'error';

const VARIANT_CLASSES: Record<MessageVariant, string> = {
  success: 'text-green-600',
  info: 'text-blue-600',
  warning: 'text-yellow-600',
  error: 'text-red-600'
};

const VARIANT_ICONS: Record<MessageVariant, IconVariant> = {
  success: 'CheckCircleIcon',
  info: 'InformationCircleIcon',
  warning: 'ExclamationCircleIcon',
  error: 'XCircleIcon'
};

export interface MessageProps {
  id: number;
  title: string;
  message?: string;
  type: MessageVariant;
  dismissAfter?: number | null;
}

export function Message({ type, title, message, dismissAfter = 2500 }: MessageProps) {
  const [show, setShow] = useState(true);
  useTimeout(() => setShow(false), dismissAfter);

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
      <div className="pointer-events-auto w-full max-w-sm overflow-hidden rounded-lg bg-white shadow-lg ring-1 ring-black ring-opacity-5">
        <div className="p-4">
          <div className="flex items-start">
            <div className="flex-shrink-0">
              <Icon
                variant={VARIANT_ICONS[type] as IconVariant}
                className={concat(VARIANT_CLASSES[type], 'h-6 w-6')}
                aria-hidden="true"
              />
            </div>
            <div className="ml-3 w-0 flex-1 pt-0.5">
              <p className="text-sm font-medium text-gray-900">{title}</p>
              <p className="mt-1 text-sm text-gray-500">{message}</p>
            </div>
            {dismissAfter !== 0 && (
              <div className="ml-4 flex flex-shrink-0">
                <Button size="sm" icon="XMarkIcon" onClick={() => setShow(false)} aria-label="Dismiss" />
              </div>
            )}
          </div>
        </div>
      </div>
    </Transition>
  );
}
