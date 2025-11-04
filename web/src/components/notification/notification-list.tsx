import { useQuery } from "@tanstack/react-query";
import { Bell } from "lucide-react";

import { NotificationItem } from "@/components/notification";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { v1NotificationsGetOptions } from "@/lib/api";

export function NotificationList() {
  const {
    data: notifications,
    isLoading,
    refetch,
  } = useQuery({
    ...v1NotificationsGetOptions(),
  });

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
      <Empty>
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Bell />
          </EmptyMedia>
          <EmptyTitle>No notifications</EmptyTitle>
          <EmptyDescription>You're all caught up!</EmptyDescription>
        </EmptyHeader>
      </Empty>
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
