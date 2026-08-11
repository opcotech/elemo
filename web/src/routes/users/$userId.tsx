import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import { AuthenticatedLayout } from "@/components/layout/authenticated-layout";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { UserAvatar } from "@/components/ui/user-avatar";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";
import { v1OrganizationsGet, v1UserGet } from "@/lib/client/sdk.gen";
import { formatDate } from "@/lib/utils";

type LoaderData = {
  user: any;
  organizations: any[];
};

export const Route = createFileRoute("/users/$userId")({
  beforeLoad: requireAuthBeforeLoad,
  loader: async ({ params }): Promise<LoaderData> => {
    // Fetch user + organizations in parallel
    const [userResult, orgsResult] = await Promise.all([
      v1UserGet({
        // NOTE: Options type flattens V1UserGetData, so we pass `path` directly.
        path: { id: params.userId },
      }),
      v1OrganizationsGet(),
    ]);

    // These shapes depend on the generated client; treating them as `{ data, error }`
    const user = (userResult as any).data ?? userResult;
    const organizations =
      ((orgsResult as any).data as any[]) ?? ((orgsResult as any[]) ?? []);

    return {
      user,
      organizations,
    };
  },
  component: UserProfilePage,
});

function UserProfilePage() {
  const { user, organizations } = Route.useLoaderData() as LoaderData;
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();

  useEffect(() => {
    setBreadcrumbsFromItems([
      { label: "Profile", href: "/profile", isNavigatable: true },
      {
        label: `${user.first_name} ${user.last_name}`,
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems, user.first_name, user.last_name]);

  // Dummy recent activity data (pure front-end)
  const recentActivity = [
    {
      id: 1,
      title: "Profile Updated",
      description: "Updated bio and contact information",
      timeAgo: "2 hours ago",
      dotClass: "bg-blue-500",
    },
    {
      id: 2,
      title: "Joined Project",
      description: "Added as team member with Developer role",
      timeAgo: "1 day ago",
      dotClass: "bg-blue-500",
    },
    {
      id: 3,
      title: "Task Completed",
      description: "Successfully completed responsive design",
      timeAgo: "3 days ago",
      dotClass: "bg-green-500",
    },
  ];

  const orgs = (organizations ?? []) as any[];

  return (
    <AuthenticatedLayout>
      <div className="container mx-auto px-4 py-8 lg:flex lg:gap-6">
        {/* Left side: profile + org memberships */}
        <div className="flex-1 space-y-4">
          <Card className="space-y-4 p-6">
            {/* Header: avatar + name + status */}
            <div className="flex flex-col gap-6 md:flex-row md:items-start">
              <div className="flex-shrink-0">
                <UserAvatar
                  firstName={user.first_name}
                  lastName={user.last_name}
                  email={user.email}
                  picture={user.picture || undefined}
                  size="lg"
                />
              </div>

              <div className="flex-1 space-y-2">
                <div className="flex flex-wrap items-center gap-2">
                  <h1 className="text-2xl font-semibold">
                    {user.first_name} {user.last_name}
                  </h1>
                  {user.status && (
                    <Badge variant="outline" className="capitalize">
                      {typeof user.status === "string"
                        ? user.status
                        : user.status.id}
                    </Badge>
                  )}
                </div>
                <p className="text-sm text-muted-foreground">@{user.username}</p>
                {user.title && <p className="text-sm">{user.title}</p>}
                {user.created_at && (
                  <p className="text-sm text-muted-foreground">
                    Member since {formatDate(user.created_at)}
                  </p>
                )}
              </div>
            </div>

            <Separator className="my-4" />

            {/* About + contact + extra details */}
            <div className="grid gap-8 md:grid-cols-2">
              <div className="space-y-3">
                <h2 className="text-sm font-medium">About</h2>
                <p className="text-sm text-muted-foreground">
                  {user.bio || "No bio added yet."}
                </p>

                <div className="space-y-1">
                  <h3 className="text-sm font-medium">Contact Information</h3>
                  <p className="text-sm">{user.email}</p>
                  {user.phone && (
                    <p className="text-sm text-muted-foreground">
                      {user.phone}
                    </p>
                  )}
                  {user.address && (
                    <p className="text-sm text-muted-foreground">
                      {user.address}
                    </p>
                  )}
                </div>
              </div>

              <div className="space-y-3">
                <h2 className="text-sm font-medium">Additional Details</h2>

                {user.languages && user.languages.length > 0 && (
                  <div className="space-y-1">
                    <p className="text-xs text-muted-foreground">Languages</p>
                    <div className="flex flex-wrap gap-2">
                      {user.languages.map((lang: string) => (
                        <Badge key={lang} variant="outline">
                          {lang}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}

                {user.links && user.links.length > 0 && (
                  <div className="space-y-1">
                    <p className="text-xs text-muted-foreground">Links</p>
                    <div className="flex flex-wrap gap-2">
                      {user.links.map((link: string) => (
                        <a
                          key={link}
                          href={link}
                          target="_blank"
                          rel="noreferrer"
                          className="text-xs font-medium underline"
                        >
                          {link}
                        </a>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          </Card>

          {/* Organization Memberships (replaces Projects & Issues) */}
          <Card>
            <CardHeader>
              <CardTitle>Organization Memberships</CardTitle>
              <CardDescription>
                Organizations and teams you&apos;re part of.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {orgs.length === 0 && (
                <p className="text-sm text-muted-foreground">
                  This user is not a member of any organizations yet.
                </p>
              )}

              {orgs.map((org) => (
                <div
                  key={org.id}
                  className="rounded-lg border bg-muted/30 p-4 text-sm"
                >
                  <div className="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
                    <div>
                      <h3 className="font-medium">{org.name}</h3>
                      {org.email && (
                        <p className="text-xs text-muted-foreground">
                          {org.email}
                        </p>
                      )}
                      {org.website && (
                        <a
                          href={org.website}
                          target="_blank"
                          rel="noreferrer"
                          className="text-xs font-medium text-primary underline"
                        >
                          {org.website}
                        </a>
                      )}
                    </div>

                    {org.status && (
                      <Badge
                        variant="outline"
                        className="mt-1 self-start capitalize"
                      >
                        {typeof org.status === "string"
                          ? org.status
                          : org.status.id}
                      </Badge>
                    )}
                  </div>

                  <div className="mt-3 flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
                    <span>
                      {(org.members?.length ?? 0).toString()} members
                    </span>
                    {org.created_at && (
                      <span>Joined {formatDate(org.created_at)}</span>
                    )}

                    <Button
                      variant="outline"
                      size="sm"
                      className="ml-auto"
                      asChild
                    >
                      {/* Replace with a real organization route if/when it exists */}
                      View Organization
                    </Button>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        </div>

        {/* Right side: Recent Activity with dummy data */}
        <div className="mt-6 w-full max-w-sm lg:mt-0">
          <Card>
            <CardHeader>
              <CardTitle>Recent Activity</CardTitle>
              <CardDescription>
                Your latest actions and contributions.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {recentActivity.map((item) => (
                <div key={item.id} className="flex gap-3">
                  <span
                    className={`mt-1 h-2 w-2 flex-shrink-0 rounded-full ${item.dotClass}`}
                  />
                  <div className="space-y-0.5">
                    <p className="text-sm font-medium">{item.title}</p>
                    <p className="text-xs text-muted-foreground">
                      {item.description}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {item.timeAgo}
                    </p>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        </div>
      </div>
    </AuthenticatedLayout>
  );
}
