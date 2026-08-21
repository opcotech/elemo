import { useQuery } from "@tanstack/react-query";
import { Bell } from "lucide-react";
import { useMemo } from "react";

import { NotificationItem } from "@/components/notification";
import {
  CursorPaginator,
  cursorPaginatorProps,
} from "@/components/shared/cursor-paginator";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { useCursorPageNav } from "@/hooks/use-cursor-page-nav";
import { cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1NotificationsGetOptions } from "@/lib/api/query-options";

export function NotificationList() {
  const pageNav = useCursorPageNav();
  const { data: notificationsPage, isLoading } = useQuery({
    ...v1NotificationsGetOptions({
      query: cursorPageQuery(pageNav.pageToken),
    }),
  });
  const notifications = notificationsPage?.items;

  const sortedNotifications = useMemo(() => {
    return notifications?.toSorted(
      (a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    );
  }, [notifications]);

  if (isLoading) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </div>
    );
  }

  if (!sortedNotifications || sortedNotifications.length === 0) {
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
          {sortedNotifications.map((notification) => (
            <NotificationItem
              key={notification.id}
              notification={notification}
            />
          ))}
          <CursorPaginator
            {...cursorPaginatorProps(notificationsPage, pageNav)}
          />
        </div>
      </ScrollArea>
    </div>
  );
}
