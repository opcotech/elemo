import { Link } from "@tanstack/react-router";
import { Edit, Plus, Shield, Trash2 } from "lucide-react";
import { useMemo, useState } from "react";

import { RoleDeleteDialog } from "./role-delete-dialog";

import { SettingsResourceTable } from "@/components/settings/settings-resource-table";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { TableSkeleton } from "@/components/ui/table-skeleton";
import {
  ResourceType,
  usePermissionsByResourceId,
} from "@/hooks/use-permissions";
import type { EffectiveActions, Role } from "@/lib/api/types";
import { Action, can } from "@/lib/auth/permissions";

const rolesListSkeletonColumns = [
  { header: "Name", skeletonClassName: "h-5 w-32" },
  { header: "Key", skeletonClassName: "h-4 w-24" },
  { header: "Description", skeletonClassName: "h-4 w-48" },
  { header: "Actions", skeletonClassName: "h-4 w-40" },
  {
    header: "Manage",
    skeletonClassName: "h-8 w-8",
    headerClassName: "text-right",
    cellClassName: "text-right",
    count: 2,
  },
] as const;

interface RoleRowProps {
  role: Role;
  permissions: EffectiveActions | undefined;
  isPermissionsLoading: boolean;
  organizationId: string;
  canManageRoles: boolean;
  onDeleteClick: (role: Role) => void;
}

function RoleRow({
  role,
  permissions,
  isPermissionsLoading,
  organizationId,
  canManageRoles,
  onDeleteClick,
}: RoleRowProps) {
  const hasRoleManagePermission =
    canManageRoles || can(permissions, Action.RoleManage);

  return (
    <TableRow>
      <TableCell className="font-medium">{role.name}</TableCell>
      <TableCell>
        <span className="font-mono text-xs">{role.key || "—"}</span>
      </TableCell>
      <TableCell>
        <span className="text-muted-foreground text-sm">
          {role.description || "—"}
        </span>
      </TableCell>
      <TableCell>
        <span className="text-muted-foreground text-xs">
          {role.actions?.length
            ? `${role.actions.length} action${role.actions.length === 1 ? "" : "s"}`
            : "—"}
        </span>
      </TableCell>
      <TableCell className="text-right">
        {isPermissionsLoading ? (
          <div className="flex justify-end gap-1">
            <Skeleton className="h-8 w-8" />
            <Skeleton className="h-8 w-8" />
          </div>
        ) : (
          <div className="flex items-center justify-end gap-x-1">
            {hasRoleManagePermission && (
              <>
                <Button
                  variant="ghost"
                  size="sm"
                  render={
                    <Link
                      to="/settings/organizations/$organizationId/roles/$roleId/edit"
                      params={{
                        organizationId,
                        roleId: role.id,
                      }}
                    />
                  }
                >
                  <Edit className="size-4" />
                  <span className="sr-only">Edit role</span>
                </Button>
                <Button
                  variant="destructive-ghost"
                  size="sm"
                  onClick={() => onDeleteClick(role)}
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
  );
}

interface RolesListProps {
  roles: Role[];
  isLoading: boolean;
  error: unknown;
  organizationId: string;
  organizationPermissions: EffectiveActions;
}

export function RolesList({
  roles,
  isLoading,
  error,
  organizationId,
  organizationPermissions,
}: RolesListProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);

  const canManageRoles = can(organizationPermissions, Action.RoleManage);
  const canViewRoles = can(organizationPermissions, Action.OrganizationRead);
  const rolePermissionsById = usePermissionsByResourceId(
    ResourceType.Role,
    roles.map((role) => role.id)
  );

  const filteredRoles = useMemo(() => {
    if (!searchTerm.trim()) return roles;
    const term = searchTerm.toLowerCase();
    return roles.filter(
      (role) =>
        role.name.toLowerCase().includes(term) ||
        role.key.toLowerCase().includes(term) ||
        (role.description && role.description.toLowerCase().includes(term))
    );
  }, [roles, searchTerm]);

  const handleDeleteClick = (role: Role) => {
    setSelectedRole(role);
    setDeleteDialogOpen(true);
  };

  const handleDeleteSuccess = () => {
    setDeleteDialogOpen(false);
    setSelectedRole(null);
  };

  const createButton = canManageRoles ? (
    <Button
      variant="outline"
      size="sm"
      render={
        <Link
          to="/settings/organizations/$organizationId/roles/new"
          params={{ organizationId }}
        />
      }
    >
      <Plus className="size-4" />
      Create Role
    </Button>
  ) : undefined;

  if (!canViewRoles && !canManageRoles) {
    return null;
  }

  return (
    <>
      <SettingsResourceTable
        dataSection="roles"
        title="Roles"
        description="Inspectable action bundles. Roles have no authority until granted on a scope."
        isLoading={isLoading}
        error={error}
        actionButton={createButton}
        search={{
          value: searchTerm,
          onChange: setSearchTerm,
          placeholder: "Search roles...",
          itemCount: roles.length,
        }}
        empty={{
          icon: <Shield />,
          title: "No roles found",
          description:
            "Roles bundle inspectable actions that can be granted on a scope.",
          action: createButton,
          searchTitle: "No roles found",
          searchDescription:
            "No roles match your search criteria. Try adjusting your search.",
          hasItems: roles.length > 0,
          hasFilteredItems: filteredRoles.length > 0,
        }}
        skeleton={<TableSkeleton columns={rolesListSkeletonColumns} />}
      >
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Key</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Actions</TableHead>
              <TableHead>
                <span className="sr-only">Manage</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredRoles.map((role) => {
              const permissionQuery = rolePermissionsById.get(role.id);
              return (
                <RoleRow
                  key={role.id}
                  role={role}
                  permissions={permissionQuery?.data}
                  isPermissionsLoading={permissionQuery?.isLoading ?? true}
                  organizationId={organizationId}
                  canManageRoles={canManageRoles}
                  onDeleteClick={handleDeleteClick}
                />
              );
            })}
          </TableBody>
        </Table>
      </SettingsResourceTable>

      {selectedRole && (
        <RoleDeleteDialog
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
