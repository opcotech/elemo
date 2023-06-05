'use client';

import { Fragment } from 'react';
import type { ReactNode } from 'react';
import { Dialog, Transition } from '@headlessui/react';

import { concat } from '@/lib/helpers';
import { Button } from '@/components/blocks/Button';
import ErrorBoundary from '@/components/ErrorBoundary';

export type Drawers = {
  default: boolean;
  todos: boolean;
  notifications: boolean;
};

export interface DrawerProps {
  id: keyof Drawers;
  title: string;
  wide?: boolean;
  children: ReactNode;
  show: boolean;
  toggle: () => void;
}

export function Drawer({ id, title, wide, children, show, toggle }: DrawerProps) {
  return (
    <Transition.Root show={Boolean(show)} as={Fragment}>
      <Dialog id={id} as="div" className="relative z-30" onClose={toggle}>
        <Transition.Child
          as={Fragment}
          enter="ease-in-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in-out duration-300"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
        </Transition.Child>

        <div className="fixed inset-0 overflow-hidden">
          <div className="absolute inset-0 overflow-hidden">
            <div className="pointer-events-none fixed inset-y-0 right-0 flex max-w-full pl-10">
              <Transition.Child
                as={Fragment}
                enter="transform transition ease-in-out duration-300 sm:duration-300"
                enterFrom="translate-x-full"
                enterTo="translate-x-0"
                leave="transform transition ease-in-out duration-300 sm:duration-300"
                leaveFrom="translate-x-0"
                leaveTo="translate-x-full"
              >
                <Dialog.Panel className={concat(wide ? 'max-w-3xl' : 'max-w-lg', 'pointer-events-auto w-screen')}>
                  <div className="flex h-full flex-col overflow-y-scroll bg-white py-6 shadow-xl">
                    <div className="px-4 sm:px-6">
                      <div className="flex items-start justify-between">
                        <Dialog.Title as={'h3'}>{title}</Dialog.Title>
                        <div className="ml-3 flex h-7 items-center">
                          <Button size="sm" icon="XMarkIcon" onClick={toggle} />
                        </div>
                      </div>
                    </div>
                    <div id={`${id}-content`} className="relative mt-6 flex-1 px-4 sm:px-6">
                      <ErrorBoundary>{children}</ErrorBoundary>
                    </div>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </div>
      </Dialog>
    </Transition.Root>
  );
}
