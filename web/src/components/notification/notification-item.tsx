import { formatDistanceToNow } from "date-fns";
import { Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useDeleteNotification } from "@/hooks/use-notifications";
import type { Notification } from "@/lib/api";

interface NotificationItemProps {
  notification: Notification;
  onSuccess?: () => void;
}

export function NotificationItem({
  notification,
  onSuccess,
}: NotificationItemProps) {
  const deleteMutation = useDeleteNotification();

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
