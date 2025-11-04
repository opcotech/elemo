import { useQuery } from "@tanstack/react-query";
import { Plus, Trash2 } from "lucide-react";
import { useState } from "react";

import { RolePermissionAddDialog } from "./role-permission-add-dialog";
import { RolePermissionDeleteDialog } from "./role-permission-delete-dialog";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
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
import type { Permission } from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import { v1OrganizationRolePermissionsGetOptions } from "@/lib/client/@tanstack/react-query.gen";
import { getFieldValue } from "@/lib/forms";

interface RolePermissionAssignmentProps {
  organizationId: string;
  roleId: string;
}

export function RolePermissionAssignment({
  organizationId,
  roleId,
}: RolePermissionAssignmentProps) {
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedPermission, setSelectedPermission] =
    useState<Permission | null>(null);

  const { data: orgPermissions, isLoading: isOrgPermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Organization, organizationId));

  const { data: rolePermissions, isLoading: isRolePermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Role, roleId));

  const hasOrgWritePermission = can(orgPermissions, "write");
  const hasRoleWritePermission = can(rolePermissions, "write");
  const hasWritePermission = hasOrgWritePermission && hasRoleWritePermission;
  const isPermissionsLoading =
    isOrgPermissionsLoading || isRolePermissionsLoading;

  const {
    data: permissions,
    isLoading,
    error,
  } = useQuery(
    v1OrganizationRolePermissionsGetOptions({
      path: {
        id: organizationId,
        role_id: roleId,
      },
    })
  );

  const extractResourceId = (target: string): string => {
    const targetValue = getFieldValue(target);
    const parts = targetValue.split(":");
    return parts.length > 1 ? parts[1] : targetValue;
  };

  const formatResourceId = (resourceId: string): string => {
    if (resourceId === "00000000000000000000") {
      return "System";
    }
    return resourceId;
  };

  const handleDeleteClick = (permission: Permission) => {
    setSelectedPermission(permission);
    setDeleteDialogOpen(true);
  };

  const handleDeleteSuccess = () => {
    setDeleteDialogOpen(false);
    setSelectedPermission(null);
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Permissions</CardTitle>
          <CardDescription>
            Manage permissions assigned to this role.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Resource Type</TableHead>
                <TableHead>Resource ID</TableHead>
                <TableHead>Permission Kind</TableHead>
                <TableHead>
                  <span className="sr-only">Actions</span>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {Array.from({ length: 3 }).map((_, i) => (
                <TableRow key={i}>
                  <TableCell>
                    <Skeleton className="h-5 w-24" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-5 w-32" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-5 w-16" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-8 w-8" />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Permissions</CardTitle>
          <CardDescription>
            Manage permissions assigned to this role.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Alert variant="destructive">
            <AlertTitle>Failed to load permissions</AlertTitle>
            <AlertDescription>
              {error instanceof Error ? error.message : "Unknown error"}
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
              <CardTitle>Permissions</CardTitle>
              <CardDescription>
                Manage permissions assigned to this role. Only
                organization-scoped resources can be assigned.
              </CardDescription>
            </div>
            {!isPermissionsLoading && hasWritePermission && (
              <Button
                type="button"
                variant="outline"
                onClick={() => setAddDialogOpen(true)}
                size="sm"
              >
                <Plus className="size-4" />
                Add Permission
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          {!permissions || permissions.length === 0 ? (
            <div className="text-muted-foreground py-8 text-center text-sm">
              No permissions assigned yet.
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Resource Type</TableHead>
                  <TableHead>Resource ID</TableHead>
                  <TableHead>Permission Kind</TableHead>
                  <TableHead>
                    <span className="sr-only">Actions</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {permissions.map((permission: Permission) => {
                  const resourceId = extractResourceId(
                    getFieldValue(permission.target)
                  );
                  return (
                    <TableRow key={permission.id}>
                      <TableCell className="font-medium">
                        {getFieldValue(permission.target_type)}
                      </TableCell>
                      <TableCell>{formatResourceId(resourceId)}</TableCell>
                      <TableCell>{permission.kind}</TableCell>
                      <TableCell className="text-right">
                        {hasWritePermission && (
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDeleteClick(permission)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <RolePermissionAddDialog
        organizationId={organizationId}
        roleId={roleId}
        open={addDialogOpen}
        onOpenChange={setAddDialogOpen}
      />

      {selectedPermission && (
        <RolePermissionDeleteDialog
          permission={selectedPermission}
          organizationId={organizationId}
          roleId={roleId}
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          onSuccess={handleDeleteSuccess}
        />
      )}
    </>
  );
}
