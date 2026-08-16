import { useMutation, useQueryClient } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import { Check, Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  v1NotificationDeleteMutation,
  v1NotificationUpdateMutation,
} from "@/lib/api/mutation-options";
import { v1NotificationsGetOptions } from "@/lib/api/query-options";
import type { Notification, NotificationPage } from "@/lib/client";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

interface NotificationItemProps {
  notification: Notification;
  onSuccess?: () => void;
}

export function NotificationItem({
  notification,
  onSuccess,
}: NotificationItemProps) {
  const queryClient = useQueryClient();
  const notificationsQueryKey = v1NotificationsGetOptions().queryKey;

  const deleteMutation = useMutation({
    ...v1NotificationDeleteMutation(),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: notificationsQueryKey });
      const previous = queryClient.getQueryData<NotificationPage>(
        notificationsQueryKey
      );
      queryClient.setQueryData<NotificationPage>(
        notificationsQueryKey,
        (current) =>
          current
            ? {
                ...current,
                items: current.items.filter(
                  (item) => item.id !== notification.id
                ),
              }
            : current
      );
      return { previous };
    },
    onSuccess: () => {
      showSuccessToast(
        "Notification deleted",
        "The notification has been removed"
      );
    },
    onError: (error, _variables, context) => {
      queryClient.setQueryData(notificationsQueryKey, context?.previous);
      showErrorToast("Failed to delete notification", error.message);
    },
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: notificationsQueryKey }),
  });
  const readMutation = useMutation({
    ...v1NotificationUpdateMutation(),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: notificationsQueryKey });
      const previous = queryClient.getQueryData<NotificationPage>(
        notificationsQueryKey
      );
      queryClient.setQueryData<NotificationPage>(
        notificationsQueryKey,
        (current) =>
          current
            ? {
                ...current,
                items: current.items.map((item) =>
                  item.id === notification.id ? { ...item, read: true } : item
                ),
              }
            : current
      );
      return { previous };
    },
    onError: (error, _variables, context) => {
      queryClient.setQueryData(notificationsQueryKey, context?.previous);
      showErrorToast("Failed to mark notification read", error.message);
    },
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: notificationsQueryKey }),
  });

  const handleDelete = () => {
    deleteMutation.mutate(
      {
        path: { id: notification.id },
      },
      {
        onSuccess: () => {
          onSuccess?.();
        },
      }
    );
  };

  const formatDate = (dateString: string) => {
    try {
      return formatDistanceToNow(new Date(dateString), { addSuffix: true });
    } catch {
      return "Unknown time";
    }
  };

  return (
    <div
      className={`group bg-background relative rounded-lg border p-4 transition-all hover:shadow-sm ${
        notification.read ? "opacity-75" : ""
      }`}
    >
      <div className="mb-3 flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-start gap-2">
            <h4
              className={`text-sm leading-tight font-medium ${
                notification.read ? "text-muted-foreground" : ""
              }`}
            >
              {notification.title}
            </h4>
          </div>
        </div>

        {!notification.read && (
          <Badge className="shrink-0 rounded px-1.5 py-0.5 text-xs">
            Unread
          </Badge>
        )}
      </div>

      {notification.description && (
        <p
          className={`text-muted-foreground mb-3 text-xs leading-relaxed ${
            notification.read ? "" : ""
          }`}
        >
          {notification.description}
        </p>
      )}

      <div className="flex items-center justify-between">
        <div className="text-muted-foreground text-xs">
          <span>{formatDate(notification.created_at)}</span>
        </div>

        <div className="flex items-center gap-1 opacity-0 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100">
          {!notification.read && (
            <Button
              size="sm"
              variant="ghost"
              onClick={() =>
                readMutation.mutate({
                  path: { id: notification.id },
                  body: { read: true },
                })
              }
              disabled={readMutation.isPending}
              className="size-7 p-0"
              title="Mark as read"
            >
              <Check className="size-4" />
            </Button>
          )}
          <Button
            size="sm"
            variant="ghost"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
            className="text-destructive hover:bg-destructive/10 hover:text-destructive size-7 p-0"
            title="Delete notification"
          >
            <Trash2 className="size-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
