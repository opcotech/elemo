'use client';

import { Fragment } from 'react';
import type { Dispatch, SetStateAction } from 'react';
import { Dialog, Transition } from '@headlessui/react';

import { concat } from '@/lib/helpers';
import { Button } from '@/components/blocks/Button';

const SIZES = {
  sm: 'w-1/5',
  md: 'w-2/5',
  lg: 'w-3/5',
  xl: 'w-4/5'
};

export interface ModalProps {
  open: boolean;
  setOpen: Dispatch<SetStateAction<boolean>>;
  size: keyof typeof SIZES;
  title: string;
  className?: string;
  modalClassName?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
}

export function Modal({
  open = false,
  setOpen,
  size = 'md',
  className,
  modalClassName,
  title,
  children,
  actions
}: ModalProps) {
  return (
    <Transition.Root show={open} as={Fragment}>
      <Dialog as="div" className={concat(modalClassName, 'relative z-40')} onClose={setOpen}>
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-200"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-out duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
        </Transition.Child>

        <div className="fixed inset-0 z-10 overflow-none">
          <div className="flex min-h-full items-center justify-center text-left sm:p-0">
            <Transition.Child
              as={Fragment}
              enter="ease-out duration-200"
              enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              enterTo="opacity-100 translate-y-0 sm:scale-100"
              leave="ease-in duration-100"
              leaveFrom="opacity-100 translate-y-0 sm:scale-100"
              leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            >
              <Dialog.Panel
                className={concat(
                  SIZES[size],
                  className,
                  'relative transform overflow-hidden rounded-lg bg-white text-left shadow-xl transition-all sm:my-8'
                )}
              >
                <div className="pt-6 pb-2">
                  <div className={concat(title ? 'justify-between' : 'justify-end', 'flex items-start px-4 sm:px-6')}>
                    <Dialog.Title className="text-lg font-medium text-gray-900">{title}</Dialog.Title>
                    <div className="ml-3 flex h-7 items-center">
                      <Button
                        id={`btn-modal-close`}
                        size="sm"
                        icon="XMarkIcon"
                        className="rounded-full text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
                        onClick={() => setOpen(false)}
                        aria-label="Close modal"
                      />
                    </div>
                  </div>

                  <div className="max-h-modal overflow-y-auto mt-4 py-4 px-4 sm:px-6">{children}</div>
                </div>

                {actions && (
                  <div className="bg-gray-50 py-3 sm:flex sm:space-x-2 justify-end px-4 sm:px-6">{actions}</div>
                )}
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition.Root>
  );
}
