import { useCallback } from 'react';
import { Notification } from '@/lib/api';
import { concat } from '@/lib/helpers';
import { Button } from '@/components/blocks/Button';
import useStore from '@/store';

export interface NotificationListItemProps extends Notification {
  loading: boolean;
  deleting: boolean;
  handleUpdateNotification: (id: string, notification: any) => void;
  handleDeleteNotification: (id: string, deleting: boolean, timer?: NodeJS.Timeout) => void;
}

export function NotificationListItem({
  id,
  title,
  description,
  read,
  loading,
  deleting,
  handleUpdateNotification,
  handleDeleteNotification
}: NotificationListItemProps) {
  const deleteNotification = useStore((state) => state.deleteNotification);
  const interactive = !deleting && !loading;

  const handleRead = useCallback(() => {
    handleUpdateNotification(id!, { read: !read });
  }, [id, read, handleUpdateNotification]);

  const handleDelete = useCallback(() => {
    const handler = setTimeout(async () => {
      await deleteNotification(id!);
      clearInterval(handler);
      handleDeleteNotification(id!, false);
    }, 5000);

    handleDeleteNotification(id!, true, handler);
  }, [id, deleteNotification, handleDeleteNotification]);

  async function handleDeleteCancel() {
    handleDeleteNotification(id!, false);
  }

  return (
    <li className={'py-4'}>
      <div className="flex items-start space-x-3">
        <div className="flex-shrink-0 items-start">
          <input
            type="checkbox"
            disabled={!interactive}
            checked={read}
            onChange={handleRead}
            className="rounded-full disabled:bg-gray-100 disabled:border-gray-300  disabled:hover:text-gray-100 disabled:hover:border-gray-300"
          />
        </div>

        <div className="min-w-0 flex-1">
          <p className="text-gray-900">
            <span className={read ? 'line-through' : ''}>{title}</span>
          </p>
          {description && <p className={concat('text-sm text-gray-500', read ? 'line-through' : '')}>{description}</p>}
        </div>
        <div className="space-x-2">
          {deleting && (
            <Button size="xs" icon="ArrowUturnLeftIcon" onClick={handleDeleteCancel}>
              <span className="sr-only">Cancel deletion</span>
            </Button>
          )}
          {!deleting && (
            <Button
              size="xs"
              icon="TrashIcon"
              onClick={handleDelete}
              disabled={!interactive}
              className="text-red-600 hover:text-red-700"
            >
              <span className="sr-only">Delete item</span>
            </Button>
          )}
        </div>
      </div>
    </li>
  );
}
