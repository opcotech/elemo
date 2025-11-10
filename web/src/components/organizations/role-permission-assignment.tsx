import { useQuery } from "@tanstack/react-query";
import { Plus, ShieldCheck, Trash2 } from "lucide-react";
import { useState } from "react";

import { RolePermissionAddDialog } from "./role-permission-add-dialog";
import { RolePermissionDeleteDialog } from "./role-permission-delete-dialog";

import { Button } from "@/components/ui/button";
import { ListContainer } from "@/components/ui/list-container";
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
import {
  extractResourceId,
  formatResourceId,
  getDefaultValue,
} from "@/lib/utils";

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

  const handleDeleteClick = (permission: Permission) => {
    setSelectedPermission(permission);
    setDeleteDialogOpen(true);
  };

  const handleDeleteSuccess = () => {
    setDeleteDialogOpen(false);
    setSelectedPermission(null);
  };

  const createButton =
    !isPermissionsLoading && hasWritePermission ? (
      <Button
        type="button"
        variant="outline"
        onClick={() => setAddDialogOpen(true)}
        size="sm"
      >
        <Plus className="size-4" />
        Add Permission
      </Button>
    ) : undefined;

  const emptyState =
    !permissions || permissions.length === 0
      ? {
          icon: <ShieldCheck />,
          title: "No permissions assigned",
          description:
            "Add permissions to grant access to organization resources.",
          action: hasWritePermission ? (
            <Button
              type="button"
              variant="outline"
              onClick={() => setAddDialogOpen(true)}
              size="sm"
            >
              <Plus className="size-4" />
              Add Permission
            </Button>
          ) : undefined,
        }
      : undefined;

  return (
    <>
      <ListContainer
        title="Permissions"
        description="Manage permissions assigned to this role. Only organization-scoped resources can be assigned."
        isLoading={isLoading}
        error={error}
        emptyState={emptyState}
        actionButton={createButton}
      >
        {isLoading ? (
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
              {permissions?.map((permission: Permission) => {
                const resourceId = extractResourceId(
                  getDefaultValue(permission.target)
                );
                return (
                  <TableRow key={permission.id}>
                    <TableCell className="font-medium">
                      {getDefaultValue(permission.target_type)}
                    </TableCell>
                    <TableCell>{formatResourceId(resourceId)}</TableCell>
                    <TableCell>{permission.kind}</TableCell>
                    <TableCell className="text-right">
                      {hasWritePermission && (
                        <Button
                          type="button"
                          variant="destructive-ghost"
                          size="sm"
                          onClick={() => handleDeleteClick(permission)}
                        >
                          <Trash2 className="h-4 w-4" />
                          <span className="sr-only">Delete permission</span>
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        )}
      </ListContainer>

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
