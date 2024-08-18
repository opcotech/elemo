import { createFileRoute } from "@tanstack/react-router";
import { User } from "lucide-react";
import { useEffect } from "react";

import { AuthenticatedLayout } from "@/components/layout/authenticated-layout";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAuth } from "@/hooks/use-auth";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/dashboard")({
  beforeLoad: requireAuthBeforeLoad,
  component: Dashboard,
});

function Dashboard() {
  const { user } = useAuth();
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();

  const getInitials = (firstName: string | null, lastName: string | null) => {
    return `${firstName?.[0] || ""}${lastName?.[0] || ""}`.toUpperCase() || "U";
  };

  useEffect(() => {
    setBreadcrumbsFromItems([
      {
        label: "Dashboard",
        isNavigatable: false,
      },
    ]);
  }, []);

  return (
    <AuthenticatedLayout>
      <div className="container mx-auto px-4 py-8">
        <div className="mb-8 flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Welcome back!</h1>
            <p className="text-muted-foreground">
              Here's what's happening with your account today.
            </p>
          </div>
        </div>

        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {/* User Profile Card */}
          <Card>
            <CardHeader className="flex flex-row items-center space-y-0 pb-2">
              <div className="flex items-center space-x-2">
                <Avatar>
                  <AvatarImage
                    src={user?.picture || undefined}
                    alt={user?.username}
                  />
                  <AvatarFallback>
                    {getInitials(
                      user?.first_name || null,
                      user?.last_name || null
                    )}
                  </AvatarFallback>
                </Avatar>
                <div>
                  <CardTitle className="text-lg">
                    {user?.first_name} {user?.last_name}
                  </CardTitle>
                  <CardDescription>@{user?.username}</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-2 text-sm">
                <div className="flex items-center gap-2">
                  <User className="text-muted-foreground h-4 w-4" />
                  <span>{user?.email}</span>
                </div>
                {user?.title && (
                  <p className="text-muted-foreground">{user.title}</p>
                )}
                {user?.bio && (
                  <p className="text-muted-foreground text-sm">{user.bio}</p>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Quick Stats */}
          <Card>
            <CardHeader>
              <CardTitle>Quick Stats</CardTitle>
              <CardDescription>Your account overview</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Status</span>
                  <span className="capitalize">{user?.status}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Member since</span>
                  <span>
                    {user?.created_at
                      ? new Date(user.created_at).toLocaleDateString()
                      : "Unknown"}
                  </span>
                </div>
                {user?.languages && user.languages.length > 0 && (
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Languages</span>
                    <span>{user.languages.join(", ")}</span>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Quick Actions */}
          <Card>
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
              <CardDescription>Common tasks</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              <Button variant="outline" className="w-full justify-start">
                View Profile
              </Button>
              <Button variant="outline" className="w-full justify-start">
                Settings
              </Button>
              <Button variant="outline" className="w-full justify-start">
                Help & Support
              </Button>
            </CardContent>
          </Card>
        </div>

        {/* Additional sections can be added here */}
        <div className="mt-8">
          <Card>
            <CardHeader>
              <CardTitle>Recent Activity</CardTitle>
              <CardDescription>Your latest actions and updates</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground py-8 text-center">
                No recent activity to display.
              </p>
            </CardContent>
          </Card>
        </div>
      </div>
    </AuthenticatedLayout>
  );
}
