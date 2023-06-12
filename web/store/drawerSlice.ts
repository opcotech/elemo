import { Drawers } from '@/components/blocks/Drawer';
import type { StateCreator } from 'zustand';

export interface DrawerSliceState {
  showing: Drawers;
  toggleDrawer: (drawer: keyof Drawers) => void;
}

export const createDrawerSlice: StateCreator<DrawerSliceState> = (set) => ({
  showing: {
    default: false,
    todos: false,
    notifications: false
  },
  toggleDrawer: (drawer) =>
    set((state) => ({
      showing: {
        ...state.showing,
        [drawer]: !state.showing[drawer]
      }
    }))
});
