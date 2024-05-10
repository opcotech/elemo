'use client';

import { useCallback, useEffect, useState } from 'react';
import useStore from '@/store';
import { Drawer } from '@/components/blocks/Drawer';
import { ListSkeleton } from '@/components/blocks/Skeleton/ListSkeleton';
import { NotificationList } from './NotificationList';
import { UpdateNotificationParams } from '@/store/notificationSlice';

type NotificationListState = {
  deleting: { id: string; timer: NodeJS.Timeout | undefined }[];
  loading: string[];
};

export function NotificationDrawer() {
  const [show, toggleDrawer, notifications, fetchingNotifications, fetchNotifications, updateNotification] = useStore(
    (state) => [
      state.showing.notifications,
      () => state.toggleDrawer('notifications'),
      state.notifications,
      state.fetchingNotifications,
      state.fetchNotifications,
      state.updateNotification
    ]
  );

  const [state, setState] = useState<NotificationListState>({
    deleting: [],
    loading: []
  });

  useEffect(() => {
    if (show && !fetchingNotifications && notifications === undefined) fetchNotifications();
  }, [show, fetchingNotifications, notifications, fetchNotifications]);

  const setLoading = useCallback(
    (id: string, loading: boolean) =>
      setState((state) => ({
        ...state,
        loading: loading ? [...state.loading, id] : state.loading.filter((l) => l !== id)
      })),
    []
  );

  const handleUpdate = useCallback(
    async (id: string, params: UpdateNotificationParams) => {
      setLoading(id, true);
      await updateNotification(id, params);
      setLoading(id, false);
    },
    [setLoading, updateNotification]
  );

  const handleDelete = useCallback(
    (id: string, deleting: boolean, timer?: NodeJS.Timeout) => {
      setState((state) => ({
        ...state,
        deleting: deleting
          ? [...state.deleting, { id, timer }]
          : state.deleting.filter((d) => {
              if (d.id === id && d.timer) clearTimeout(d.timer);
              return d.id !== id;
            })
      }));
      setLoading(id, deleting);
    },
    [setLoading]
  );

  function handleDrawerClose() {
    toggleDrawer();
  }

  return (
    <Drawer id="notifications" title="Notifications" show={show} toggle={handleDrawerClose}>
      <div className="px-3">
        {fetchingNotifications && notifications === undefined ? (
          <ListSkeleton count={5} fullWidth />
        ) : (
          <NotificationList
            notifications={notifications || []}
            deleting={state.deleting}
            loading={state.loading}
            handleUpdate={handleUpdate}
            handleDelete={handleDelete}
          />
        )}
      </div>
    </Drawer>
  );
}
