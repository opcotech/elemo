'use client';

import Drawer from '@/components/Drawer';
import {ContentSkeleton} from '@/components/Skeleton';
import useStore from '@/store';

export default function TodoDrawer() {
  const [showTodos, toggleDrawer] = useStore((state) => [state.drawers.showTodos, state.toggleDrawer]);

  return (
    <Drawer
      id="showTodos"
      title="Todos"
      show={showTodos}
      toggle={() => toggleDrawer('showTodos')}
    >
      <ContentSkeleton/>
    </Drawer>
  );
}
