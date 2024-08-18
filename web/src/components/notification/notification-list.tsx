import { Bell } from "lucide-react";

import { NotificationItem } from "@/components/notification";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { useNotifications } from "@/hooks/use-notifications";

export function NotificationList() {
  const { data: notifications, isLoading, refetch } = useNotifications();

  if (isLoading) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </div>
    );
  }

  if (!notifications || notifications.length === 0) {
    return (
      <div className="flex h-32 flex-col items-center justify-center space-y-2">
        <Bell className="text-muted-foreground h-8 w-8" />
        <p className="text-muted-foreground text-sm">No notifications</p>
        <p className="text-muted-foreground text-xs">You're all caught up!</p>
      </div>
    );
  }

  return (
    <div className="h-full">
      <ScrollArea className="h-full">
        <div className="space-y-3 pr-2 pb-4">
          {notifications.map((notification) => (
            <NotificationItem
              key={notification.id}
              notification={notification}
              onSuccess={() => {
                refetch();
              }}
            />
          ))}
        </div>
      </ScrollArea>
    </div>
  );
}
