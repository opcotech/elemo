'use client';

import Drawer from '@/components/Drawer';
import {ContentSkeleton} from '@/components/Skeleton';
import useStore from '@/store';

export default function NotificationDrawer() {
  const [showNotifications, toggleDrawer] = useStore((state) => [state.drawers.showNotifications, state.toggleDrawer]);

  return (
    <Drawer
      id="showNotifications"
      title="Notifications"
      show={showNotifications}
      toggle={() => toggleDrawer('showNotifications')}
    >
      <ContentSkeleton />
    </Drawer>
  );
}
