'use client';

import useStore from '@/store';
import { Drawer } from '@/components/blocks/Drawer';

export interface NotificationDrawerProps {}

export function NotificationDrawer({}: NotificationDrawerProps) {
  const [show, toggleDrawer] = useStore((state) => [
    state.showing.notifications,
    () => state.toggleDrawer('notifications')
  ]);

  function handleDrawerClose() {
    toggleDrawer();
  }

  return (
    <Drawer id="notifications" title="Notifications" show={show} toggle={handleDrawerClose}>
      <p>Notifications</p>
    </Drawer>
  );
}
