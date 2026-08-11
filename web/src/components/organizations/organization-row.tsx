import { Link } from "@tanstack/react-router";
import { Edit, Trash2 } from "lucide-react";
import { useState } from "react";

import { OrganizationDeleteDialog } from "./organization-delete-dialog";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ConditionalLink } from "@/components/ui/conditional-link";
import { ExternalLink } from "@/components/ui/external-link";
import { Skeleton } from "@/components/ui/skeleton";
import { TableCell, TableRow } from "@/components/ui/table";
import type { Organization, Permission } from "@/lib/api/types";
import { can } from "@/lib/auth/permissions";
import { zOrganizationStatus } from "@/lib/client/zod.gen";
import { pluralize } from "@/lib/utils";

export function OrganizationRow({
  organization,
  permissions,
  isPermissionsLoading,
}: {
  organization: Organization;
  permissions?: Permission[];
  isPermissionsLoading: boolean;
}) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const hasReadPermission = can(permissions, "read");
  const hasWritePermission = can(permissions, "write");
  const hasDeletePermission = can(permissions, "delete");

  return (
    <TableRow>
      <TableCell className="font-medium">
        <ConditionalLink
          to="/settings/organizations/$organizationId"
          params={{ organizationId: organization.id }}
          condition={hasReadPermission}
        >
          {organization.name}
        </ConditionalLink>
      </TableCell>
      <TableCell>{organization.email}</TableCell>
      <TableCell>
        {organization.website ? (
          <ExternalLink href={organization.website} />
        ) : (
          <span className="text-muted-foreground">—</span>
        )}
      </TableCell>
      <TableCell>
        <Badge variant="secondary">
          {organization.members.length}{" "}
          {pluralize(organization.members.length, "member", "members")}
        </Badge>
      </TableCell>
      <TableCell>
        {organization.status === zOrganizationStatus.enum.active ? (
          <Badge variant="success">Active</Badge>
        ) : (
          <Badge variant="destructive">Deleted</Badge>
        )}
      </TableCell>
      <TableCell className="text-right">
        <div className="flex items-center justify-end gap-x-1">
          {isPermissionsLoading ? (
            <div className="flex items-center gap-x-2 py-1.5">
              <Skeleton className="h-5 w-8" />
              <Skeleton className="h-5 w-8" />
              <Skeleton className="h-5 w-8" />
            </div>
          ) : (
            <>
              {hasWritePermission && (
                <Button
                  variant="ghost"
                  size="sm"
                  render={
                    <Link
                      to="/settings/organizations/$organizationId/edit"
                      params={{ organizationId: organization.id }}
                    />
                  }
                >
                  <Edit className="size-4" />
                  <span className="sr-only">Edit organization</span>
                </Button>
              )}
              {hasDeletePermission &&
                organization.status === zOrganizationStatus.enum.active && (
                  <>
                    <Button
                      variant="destructive-ghost"
                      size="sm"
                      onClick={() => setDeleteDialogOpen(true)}
                    >
                      <Trash2 className="size-4" />
                      <span className="sr-only">Delete organization</span>
                    </Button>
                    <OrganizationDeleteDialog
                      organization={organization}
                      open={deleteDialogOpen}
                      onOpenChange={setDeleteDialogOpen}
                    />
                  </>
                )}
            </>
          )}
        </div>
      </TableCell>
    </TableRow>
  );
}
