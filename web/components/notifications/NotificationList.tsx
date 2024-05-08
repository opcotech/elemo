import { memo } from 'react';
import { Notification } from '@/lib/api';
import { NotificationListItem } from './NotificationListItem';
import { UpdateNotificationParams } from '@/store/notificationSlice';

const MemoizedNotificationListItem = memo(NotificationListItem);

export interface NotificationListProps {
  notifications: Notification[];
  deleting: { id: string; timer: NodeJS.Timeout | undefined }[];
  loading: string[];
  handleUpdate: (id: string, notification: UpdateNotificationParams) => Promise<void>;
  handleDelete: (id: string, deleting: boolean, timer?: NodeJS.Timeout) => void;
}

export function NotificationList({
  notifications,
  deleting,
  loading,
  handleUpdate,
  handleDelete
}: NotificationListProps) {
  return (
    <ul>
      {notifications.length === 0 && <li className="text-center text-gray-500">No notifications found.</li>}

      {notifications.map((notification) => (
        <MemoizedNotificationListItem
          key={notification.id}
          {...notification}
          deleting={deleting.filter((d) => d.id === notification.id).length > 0}
          loading={loading.includes(notification.id)}
          handleUpdateNotification={handleUpdate}
          handleDeleteNotification={handleDelete}
        />
      ))}
    </ul>
  );
}
