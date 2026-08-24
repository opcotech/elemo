import { useQuery, useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Bell, CheckCircle2Icon } from "lucide-react";
import { useMemo } from "react";

import { NotificationList } from "@/components/notification";
import { MockDataAlert } from "@/components/shared/app-feedback";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/ui/empty-state";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CompactWorkList } from "@/components/work/work-list";
import { useAuth } from "@/hooks/use-auth";
import { collectedListQuery, cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1NotificationsGetOptions,
  v1UsersIssuesGetOptions,
} from "@/lib/api/query-options";
import { v1UsersIssuesGet } from "@/lib/api/sdk";
import { selectAttentionSignals } from "@/lib/mock-data";
import { withRouteErrors } from "@/lib/route-errors";
import { issuesToWorkItems } from "@/lib/work/issue-adapter";
import { queryWorkItems } from "@/lib/work/query";

export const Route = createFileRoute("/_authenticated/inbox")({
  staticData: {
    breadcrumb: "Inbox",
  },
  loader: ({ context }) =>
    withRouteErrors(() =>
      context.queryClient.fetchQuery(
        v1NotificationsGetOptions({
          query: cursorPageQuery(),
        })
      )
    ),
  component: InboxPage,
});

function InboxPage() {
  const { data: notificationsPage } = useSuspenseQuery(
    v1NotificationsGetOptions({
      query: cursorPageQuery(),
    })
  );
  const notifications = notificationsPage?.items;

  const unreadCount = notifications?.filter((n) => !n.read).length || 0;

  return (
    <div className="flex h-full">
      <div className="w-full lg:hidden">
        <Tabs defaultValue="inbox" className="h-full w-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="inbox">Inbox</TabsTrigger>
            <TabsTrigger value="notifications">
              Notifications
              {unreadCount > 0 && (
                <Badge variant="primary" className="ml-2">
                  {unreadCount}
                </Badge>
              )}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="inbox" className="h-full overflow-auto p-4">
            <InboxContent />
          </TabsContent>

          <TabsContent value="notifications" className="h-full overflow-auto">
            <NotificationsPanel unreadCount={unreadCount} />
          </TabsContent>
        </Tabs>
      </div>

      <div className="hidden h-full w-full lg:flex">
        <div className="flex-1 overflow-auto border-r">
          <div className="p-6">
            <InboxContent />
          </div>
        </div>

        <div className="flex h-full w-96 flex-col">
          <NotificationsPanel unreadCount={unreadCount} />
        </div>
      </div>
    </div>
  );
}

function InboxContent() {
  const { user } = useAuth();
  const userId = user?.id;
  const userIssuesOptions = v1UsersIssuesGetOptions({
    path: { id: userId ?? "" },
    query: cursorPageQuery(),
  });
  const { data: issuesPage } = useQuery({
    ...collectedListQuery(userIssuesOptions, async (pageToken, signal) => {
      const { data } = await v1UsersIssuesGet({
        path: { id: userId ?? "" },
        query: cursorPageQuery(pageToken),
        signal,
        throwOnError: true,
      });
      return data;
    }),
    enabled: Boolean(userId),
  });
  const userWorkItems = useMemo(
    () => issuesToWorkItems(issuesPage?.items ?? []),
    [issuesPage?.items]
  );
  const userWorkItemById = useMemo(
    () =>
      new Map(
        userWorkItems.flatMap((item) => [
          [item.id, item] as const,
          [item.key, item] as const,
        ])
      ),
    [userWorkItems]
  );
  const attentionItems = useMemo(() => {
    const items: typeof userWorkItems = [];
    for (const signal of selectAttentionSignals()) {
      const item = userWorkItemById.get(signal.workItemId);
      if (item) {
        items.push(item);
      }
    }
    return items;
  }, [userWorkItemById]);
  const watched = useMemo(
    () =>
      queryWorkItems(userWorkItems, {
        filters: { statuses: ["blocked", "in progress"] },
      }),
    [userWorkItems]
  );

  return (
    <>
      <div className="mb-6">
        <h1 className="text-2xl font-semibold tracking-tight">Inbox</h1>
        <p className="text-muted-foreground">
          Notifications and operational signals that may need your response.
        </p>
      </div>

      <MockDataAlert title="Illustrative operational inbox" className="mb-4">
        Attention and watched-work tabs include illustrative examples.
        Notifications reflect your live inbox.
      </MockDataAlert>

      <Tabs defaultValue="attention" className="w-full">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="attention">Attention</TabsTrigger>
          <TabsTrigger value="watched">Watched work</TabsTrigger>
          <TabsTrigger value="handled">Handled</TabsTrigger>
        </TabsList>

        <TabsContent value="attention" className="mt-6">
          <CompactWorkList
            items={attentionItems}
            emptyTitle="Nothing needs attention"
            emptyDescription="New fixture attention signals will appear here."
          />
        </TabsContent>

        <TabsContent value="watched" className="mt-6">
          <CompactWorkList items={watched} />
        </TabsContent>

        <TabsContent value="handled" className="mt-6">
          <EmptyState
            compact
            icon={<CheckCircle2Icon />}
            title="No handled items"
            description="Acknowledged fixture signals would remain available here."
          />
        </TabsContent>
      </Tabs>
    </>
  );
}

function NotificationsPanel({ unreadCount }: { unreadCount: number }) {
  return (
    <>
      <div className="shrink-0 border-b p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Bell className="h-5 w-5" />
            <h2 className="text-lg font-semibold">Notifications</h2>
            {unreadCount > 0 && <Badge variant="primary">{unreadCount}</Badge>}
          </div>
        </div>
        <p className="text-muted-foreground mt-1 text-sm">
          Your in-app notifications
        </p>
      </div>

      <div className="flex-1 overflow-hidden p-4">
        <NotificationList />
      </div>
    </>
  );
}
