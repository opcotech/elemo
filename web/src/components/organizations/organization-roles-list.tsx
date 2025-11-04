import { Link } from "@tanstack/react-router";
import { Edit, Plus, Trash2 } from "lucide-react";
import { useState } from "react";

import { OrganizationRoleDeleteDialog } from "./organization-role-delete-dialog";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import type { Role } from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import { pluralize } from "@/lib/utils";

function OrganizationRolesListSkeleton() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Description</TableHead>
          <TableHead>Members</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 5 }).map((_, i) => (
          <TableRow key={i}>
            <TableCell>
              <Skeleton className="h-5 w-32" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-48" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-6 w-16" />
            </TableCell>
            <TableCell className="text-right">
              <div className="flex justify-end gap-1">
                <Skeleton className="h-8 w-8" />
                <Skeleton className="h-8 w-8" />
              </div>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

interface OrganizationRolesListProps {
  roles: Role[];
  isLoading: boolean;
  error: unknown;
  organizationId: string;
}

export function OrganizationRolesList({
  roles,
  isLoading,
  error,
  organizationId,
}: OrganizationRolesListProps) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);

  // Check permissions for organization (write) and role (write)
  const { data: orgPermissions, isLoading: isOrgPermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Organization, organizationId));

  const { data: rolePermissions, isLoading: isRolePermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Role, ""));

  const hasOrgWritePermission = can(orgPermissions, "write");
  const hasRoleCreatePermission = can(rolePermissions, "create");
  const hasRoleWritePermission = can(rolePermissions, "write");
  const hasWritePermission = hasOrgWritePermission && hasRoleWritePermission;
  const hasCreatePermission = hasOrgWritePermission && hasRoleCreatePermission;
  const isPermissionsLoading =
    isOrgPermissionsLoading || isRolePermissionsLoading;

  const handleDeleteClick = (role: Role) => {
    setSelectedRole(role);
    setDeleteDialogOpen(true);
  };

  const handleDeleteSuccess = () => {
    setDeleteDialogOpen(false);
    setSelectedRole(null);
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Roles</CardTitle>
          <CardDescription>
            Organization roles and their members.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <OrganizationRolesListSkeleton />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Roles</CardTitle>
          <CardDescription>
            Organization roles and their members.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Alert variant="destructive">
            <AlertDescription>
              Failed to load organization roles. Please try again later.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between">
            <div>
              <CardTitle>Roles</CardTitle>
              <CardDescription>
                Organization roles and their members.
              </CardDescription>
            </div>
            {!isPermissionsLoading && hasCreatePermission && (
              <Button variant="outline" size="sm" asChild>
                <Link
                  to="/settings/organizations/$organizationId/roles/new"
                  params={{ organizationId }}
                >
                  <Plus className="size-4" />
                  Create Role
                </Link>
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Members</TableHead>
                <TableHead>
                  <span className="sr-only">Actions</span>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {roles.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="py-4 text-center">
                    No roles found in this organization.
                  </TableCell>
                </TableRow>
              ) : (
                roles.map((role) => (
                  <TableRow key={role.id}>
                    <TableCell className="font-medium">{role.name}</TableCell>
                    <TableCell>
                      <span className="text-muted-foreground text-sm">
                        {role.description || "—"}
                      </span>
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary">
                        {role.members.length}{" "}
                        {pluralize(role.members.length, "member", "members")}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      {isPermissionsLoading ? (
                        <div className="flex justify-end gap-1">
                          <div className="h-8 w-8" />
                          <div className="h-8 w-8" />
                        </div>
                      ) : (
                        <div className="flex items-center justify-end gap-x-1">
                          {hasWritePermission && (
                            <>
                              <Button variant="ghost" size="sm" asChild>
                                <Link
                                  to="/settings/organizations/$organizationId/roles/$roleId/edit"
                                  params={{
                                    organizationId,
                                    roleId: role.id,
                                  }}
                                >
                                  <Edit className="size-4" />
                                  <span className="sr-only">Edit role</span>
                                </Link>
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleDeleteClick(role)}
                              >
                                <Trash2 className="size-4" />
                                <span className="sr-only">Delete role</span>
                              </Button>
                            </>
                          )}
                        </div>
                      )}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {selectedRole && (
        <OrganizationRoleDeleteDialog
          role={selectedRole}
          organizationId={organizationId}
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          onSuccess={handleDeleteSuccess}
        />
      )}
    </>
  );
}
