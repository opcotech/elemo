/**
 * The store intentionally have no slices for resources fetched using Apollo.
 *
 * We are using ApolloClient to manage todos, which has its own state
 * management system with caching. We are using and updating the ApolloClient
 * cache upon mutations.
 */

import {create} from 'zustand';
import {devtools} from 'zustand/middleware';

import {drawerSlice, type DrawerSliceState} from './store.drawers';
import {messageSlice, type MessageSliceState} from './store.messages';

export type StoreState = MessageSliceState & DrawerSliceState;

export const useStore = create<StoreState, [['zustand/devtools', never]]>(
  devtools((...ctx) => ({
    ...drawerSlice(...ctx),
    ...messageSlice(...ctx)
  }))
);
