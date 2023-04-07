import type {StateCreator} from 'zustand';

export interface DrawerSliceState {
  drawers: Drawers;
  toggleDrawer: (drawer: keyof Drawers) => void;
}

export const drawerSlice: StateCreator<DrawerSliceState> = (set) => ({
  drawers: {
    showTodos: false,
    showNotifications: false
  },
  toggleDrawer: (drawer) =>
    set((state) => ({
      drawers: {
        ...state.drawers,
        [drawer]: !state.drawers[drawer]
      }
    }))
});
