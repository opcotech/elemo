import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Bell, CheckCircle2Icon } from "lucide-react";

import { NotificationList } from "@/components/notification";
import { AppEmptyState, MockDataAlert } from "@/components/shared/app-feedback";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CompactWorkList } from "@/components/work/work-list";
import { v1NotificationsGetOptions } from "@/lib/api/query-options";
import {
  getWorkItem,
  selectAttentionSignals,
  selectWorkItems,
} from "@/lib/mock-data";

export const Route = createFileRoute("/_authenticated/inbox")({
  staticData: {
    breadcrumb: "Inbox",
  },
  loader: ({ context }) =>
    context.queryClient.fetchQuery(v1NotificationsGetOptions()),
  component: InboxPage,
});

function InboxPage() {
  const { data: notifications } = useSuspenseQuery(v1NotificationsGetOptions());

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
  const attentionItems = selectAttentionSignals()
    .map((signal) => getWorkItem(signal.workItemId))
    .filter((item) => item !== undefined);
  const watched = selectWorkItems({
    filters: { statuses: ["blocked", "in-progress"] },
  });

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
          <AppEmptyState
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
