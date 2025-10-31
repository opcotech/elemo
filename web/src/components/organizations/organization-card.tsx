import { Link } from "@tanstack/react-router";
import { Edit, Eye, Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { ExternalLink } from "@/components/ui/external-link";
import { Skeleton } from "@/components/ui/skeleton";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import type { Organization } from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import { formatDate, pluralize } from "@/lib/utils";

export function OrganizationCard({
  organization,
}: {
  organization: Organization;
}) {
  const { data: permissions, isLoading: isPermissionsLoading } = usePermissions(
    withResourceType(ResourceType.Organization, organization.id)
  );

  const hasReadPermission = can(permissions, "read");
  const hasWritePermission = can(permissions, "write");
  const hasDeletePermission = can(permissions, "delete");

  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="text-lg">
              <Link
                to="/settings/organizations/$organizationId"
                params={{ organizationId: organization.id }}
                className="text-primary hover:underline"
              >
                {organization.name}
              </Link>
            </CardTitle>
            <CardDescription className="mt-1">
              {organization.email}
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            {organization.status === "active" ? (
              <Badge variant="success">Active</Badge>
            ) : (
              <Badge variant="destructive">Deleted</Badge>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex items-center gap-4 text-sm">
          {organization.website && (
            <div className="flex items-center gap-1">
              <ExternalLink href={organization.website} />
            </div>
          )}
          <div className="text-muted-foreground">
            {organization.members.length}{" "}
            {pluralize(organization.members.length, "member", "members")}
          </div>
          <div className="text-muted-foreground">
            Created {formatDate(organization.created_at)}
          </div>
        </div>

        {isPermissionsLoading ? (
          <div className="flex items-center gap-x-2 pt-2">
            <Skeleton className="h-8 w-8" />
            <Skeleton className="h-8 w-8" />
            <Skeleton className="h-8 w-8" />
          </div>
        ) : (
          <div className="flex items-center gap-x-1 pt-2">
            {hasReadPermission && (
              <Button variant="ghost" size="sm" asChild>
                <Link
                  to="/settings/organizations/$organizationId"
                  params={{ organizationId: organization.id }}
                >
                  <Eye className="size-4" />
                  <span className="sr-only">View organization</span>
                </Link>
              </Button>
            )}
            {hasWritePermission && (
              <Button variant="ghost" size="sm" disabled>
                <Edit className="size-4" />
                <span className="sr-only">Edit organization</span>
              </Button>
            )}
            {hasDeletePermission && (
              <Button variant="ghost" size="sm" disabled>
                <Trash2 className="size-4" />
                <span className="sr-only">Delete organization</span>
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
