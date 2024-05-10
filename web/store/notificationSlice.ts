import { StateCreator } from 'zustand/esm';
import { getErrorMessage, Notification, NotificationService } from '@/lib/api';
import { MessageSliceState } from '@/store/messageSlice';

export type FetchNotificationParams = {
  offset?: number;
  limit?: number;
};

export type UpdateNotificationParams = {
  read: boolean;
};

export interface NotificationSliceState {
  notifications: Notification[] | undefined;
  fetchingNotifications: boolean;
  fetchNotifications: (params?: FetchNotificationParams) => Promise<void>;
  updateNotification: (id: string, params: UpdateNotificationParams) => Promise<void>;
  deleteNotification: (id: string) => Promise<void>;
}

export function sortNotifications(items: Notification[]): Notification[] {
  return Object.assign([] as Notification[], items).sort((a, b) => {
    return a.read && !b.read ? 1 : -1;
  });
}

export const createNotificationSlice: StateCreator<NotificationSliceState & Partial<MessageSliceState>> = (
  set,
  get
) => ({
  notifications: undefined,
  fetchingNotifications: false,
  fetchNotifications: async ({ offset = 0, limit = 100 }: FetchNotificationParams = {}) => {
    let notifications: Notification[] = [];

    try {
      set({ fetchingNotifications: true });
      notifications = await NotificationService.v1NotificationsGet(offset, limit);
    } catch (e) {
      return get().addMessage?.({ type: 'error', title: 'Failed to fetch notifications', message: getErrorMessage(e) });
    } finally {
      set({ notifications: sortNotifications(notifications), fetchingNotifications: false });
    }
  },
  updateNotification: async (id: string, params: UpdateNotificationParams) => {
    let updated: Notification;

    try {
      updated = await NotificationService.v1NotificationUpdate(id, params);
      get().addMessage?.({
        type: 'success',
        title: 'Notification updated',
        message: `Notification "${id}" updated successfully.`
      });
    } catch (e) {
      return get().addMessage?.({ type: 'error', title: 'Failed to update notification', message: getErrorMessage(e) });
    }

    set((state) => ({
      notifications: sortNotifications((state.notifications || []).map((todo) => (todo.id === id ? updated : todo)))
    }));
  },
  deleteNotification: async (id: string) => {
    try {
      await NotificationService.v1NotificationDelete(id);
      get().addMessage?.({
        type: 'success',
        title: 'Notification deleted',
        message: `Notification "${id}" deleted successfully.`
      });
    } catch (e) {
      return get().addMessage?.({ type: 'error', title: 'Failed to delete notification', message: getErrorMessage(e) });
    }

    set((state) => ({
      notifications: sortNotifications((state.notifications || []).filter((todo) => todo.id !== id))
    }));
  }
});
