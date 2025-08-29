import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Bell, Eye, Folder, GitBranch } from "lucide-react";
import { useEffect } from "react";

import { AuthenticatedLayout } from "@/components/layout/authenticated-layout";
import { NotificationList } from "@/components/notification";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { v1NotificationsGetOptions } from "@/lib/api";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/inbox")({
  beforeLoad: requireAuthBeforeLoad,
  component: () => (
    <AuthenticatedLayout>
      <InboxPage />
    </AuthenticatedLayout>
  ),
});

function InboxPage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const { data: notifications } = useQuery({
    ...v1NotificationsGetOptions(),
  });

  useEffect(() => {
    setBreadcrumbsFromItems([
      {
        label: "Inbox",
        isNavigatable: false,
      },
    ]);
  }, []);

  const unreadCount = notifications?.filter((n) => !n.read).length || 0;

  return (
    <div className="flex h-full">
      {/* Small screens: Tab view */}
      <div className="w-full lg:hidden">
        <Tabs defaultValue="inbox" className="w-full h-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="inbox">Inbox</TabsTrigger>
            <TabsTrigger value="notifications">
              Notifications
              {unreadCount > 0 && (
                <Badge variant="default" className="ml-2 bg-blue-500 text-white">
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

      {/* Big screens: Two-column layout */}
      <div className="hidden lg:flex h-full w-full">
        {/* Left side - Inbox */}
        <div className="flex-1 overflow-auto border-r">
          <div className="p-6">
            <InboxContent />
          </div>
        </div>

        {/* Right side - Notifications */}
        <div className="flex h-full w-96 flex-col">
          <NotificationsPanel unreadCount={unreadCount} />
        </div>
      </div>
    </div>
  );
}

function InboxContent() {
  return (
    <>
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Inbox</h1>
        <p className="text-muted-foreground">
          Stay updated with your project activities and notifications
        </p>
      </div>

      <Tabs defaultValue="projects" className="w-full">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="projects" className="flex items-center gap-2">
            <Folder className="h-4 w-4" />
            Projects
          </TabsTrigger>
          <TabsTrigger value="workspace" className="flex items-center gap-2">
            <GitBranch className="h-4 w-4" />
            Workspace
          </TabsTrigger>
          <TabsTrigger value="watched" className="flex items-center gap-2">
            <Eye className="h-4 w-4" />
            Watched
          </TabsTrigger>
        </TabsList>

        <TabsContent value="projects" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                Project Activities
                <Badge variant="secondary" className="ml-auto">
                  Coming Soon
                </Badge>
              </CardTitle>
              <CardDescription>
                Recent activities from your projects
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex h-32 items-center justify-center">
                <p className="text-muted-foreground text-sm">
                  Project activities will be displayed here
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="workspace" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                Workspace Activities
                <Badge variant="secondary" className="ml-auto">
                  Coming Soon
                </Badge>
              </CardTitle>
              <CardDescription>Activities from your workspace</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex h-32 items-center justify-center">
                <p className="text-muted-foreground text-sm">
                  Workspace activities will be displayed here
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="watched" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                Watched Issues
                <Badge variant="secondary" className="ml-auto">
                  Coming Soon
                </Badge>
              </CardTitle>
              <CardDescription>
                Updates from issues you're watching
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex h-32 items-center justify-center">
                <p className="text-muted-foreground text-sm">
                  Watched issues will be displayed here
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </>
  );
}

function NotificationsPanel({ unreadCount }: { unreadCount: number }) {
  return (
    <>
      <div className="flex-shrink-0 border-b p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Bell className="h-5 w-5" />
            <h2 className="text-lg font-semibold">Notifications</h2>
            {unreadCount > 0 && (
              <Badge variant="default" className="bg-blue-500 text-white">
                {unreadCount}
              </Badge>
            )}
          </div>
          <Button variant="ghost" size="sm">
            Mark all read
          </Button>
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
