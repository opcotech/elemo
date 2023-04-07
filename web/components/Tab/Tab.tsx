'use client';

import {Tab as HeadlessTab} from '@headlessui/react';
import type {ReactNode} from 'react';

import {concat} from '@/helpers';

export interface TabItem {
  id: string;
  name: string;
  component: ReactNode;
}

export interface TabProps {
  tabs: TabItem[];
}

export default function Tab({tabs}: TabProps) {
  return (
    <HeadlessTab.Group>
      <HeadlessTab.List className="block sm:flex mb-6 space-x-2">
        {tabs.map((tab) => (
          <HeadlessTab
            id={`settings-page-tab-${tab.id}`}
            key={tab.name}
            className={({selected}) =>
              concat(
                selected ? 'bg-gray-100 text-gray-700' : 'text-gray-500 hover:bg-gray-100 hover:text-gray-700',
                'px-3 py-2 font-medium text-sm rounded'
              )
            }
          >
            {tab.name}
          </HeadlessTab>
        ))}
      </HeadlessTab.List>
      <HeadlessTab.Panels className="px-1 py-5">
        {tabs.map((tab) => (
          <HeadlessTab.Panel key={tab.name}>{tab.component}</HeadlessTab.Panel>
        ))}
      </HeadlessTab.Panels>
    </HeadlessTab.Group>
  );
}
