import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

import { createDrawerSlice, type DrawerSliceState } from './drawerSlice';
import { createMessageSlice, type MessageSliceState } from './messageSlice';
import { createTodoSlice, type TodoSliceState } from './todoSlice';
import { createOrganizationSlice, OrganizationState } from './organizationSlice';

export type StoreState = MessageSliceState & DrawerSliceState & TodoSliceState & OrganizationState;

export const useStore = create<StoreState, [['zustand/devtools', never]]>(
  devtools((...ctx) => ({
    ...createDrawerSlice(...ctx),
    ...createMessageSlice(...ctx),
    ...createTodoSlice(...ctx),
    ...createOrganizationSlice(...ctx)
  }))
);
