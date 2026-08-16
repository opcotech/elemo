import { Link } from "@tanstack/react-router";
import { Edit, Plus, Shield, Trash2, UserPlus } from "lucide-react";
import { useMemo, useState } from "react";

import { RoleDeleteDialog } from "./role-delete-dialog";
import { RoleMemberAddDialog } from "./role-member-add-dialog";

import { SettingsResourceTable } from "@/components/settings/settings-resource-table";
import { Button } from "@/components/ui/button";
import { CountBadge } from "@/components/ui/count-badge";
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
import type { Permission, Role } from "@/lib/api/types";
import { can } from "@/lib/auth/permissions";

const rolesListSkeletonColumns = [
  { header: "Name", skeletonClassName: "h-5 w-32" },
  { header: "Description", skeletonClassName: "h-4 w-48" },
  { header: "Members", skeletonClassName: "h-6 w-16" },
  {
    header: "Actions",
    skeletonClassName: "h-8 w-8",
    headerClassName: "text-right",
    cellClassName: "text-right",
    count: 2,
  },
] as const;

interface RoleRowProps {
  role: Role;
  permissions: Permission[] | undefined;
  isPermissionsLoading: boolean;
  organizationId: string;
  onAddMemberClick: (role: Role) => void;
  onDeleteClick: (role: Role) => void;
}

function RoleRow({
  role,
  permissions,
  isPermissionsLoading,
  organizationId,
  onAddMemberClick,
  onDeleteClick,
}: RoleRowProps) {
  const hasRoleWritePermission = can(permissions, "write");
  const hasRoleDeletePermission = can(permissions, "delete");

  return (
    <TableRow>
      <TableCell className="font-medium">{role.name}</TableCell>
      <TableCell>
        <span className="text-muted-foreground text-sm">
          {role.description || "—"}
        </span>
      </TableCell>
      <TableCell>
        <CountBadge
          count={role.member_count ?? 0}
          singular="member"
          plural="members"
        />
      </TableCell>
      <TableCell className="text-right">
        {isPermissionsLoading ? (
          <div className="flex justify-end gap-1">
            <Skeleton className="h-8 w-8" />
            <Skeleton className="h-8 w-8" />
            <Skeleton className="h-8 w-8" />
          </div>
        ) : (
          <div className="flex items-center justify-end gap-x-1">
            {hasRoleWritePermission && (
              <>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onAddMemberClick(role)}
                >
                  <UserPlus className="size-4" />
                  <span className="sr-only">Add member</span>
                </Button>
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
              </>
            )}
            {hasRoleDeletePermission && (
              <Button
                variant="destructive-ghost"
                size="sm"
                onClick={() => onDeleteClick(role)}
              >
                <Trash2 className="size-4" />
                <span className="sr-only">Delete role</span>
              </Button>
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
  organizationPermissions: Permission[];
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
  const [addMemberDialogOpen, setAddMemberDialogOpen] = useState(false);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);

  const hasOrgWritePermission = can(organizationPermissions, "write");
  const hasCreatePermission = hasOrgWritePermission;
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

  const handleAddMemberClick = (role: Role) => {
    setSelectedRole(role);
    setAddMemberDialogOpen(true);
  };

  const createButton = hasCreatePermission ? (
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

  return (
    <>
      <SettingsResourceTable
        dataSection="roles"
        title="Roles"
        description="Organization roles and their members."
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
            "Roles help organize permissions and member access. Create a role to get started.",
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
              <TableHead>Description</TableHead>
              <TableHead>Members</TableHead>
              <TableHead>
                <span className="sr-only">Actions</span>
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
                  onAddMemberClick={handleAddMemberClick}
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

      {selectedRole && (
        <RoleMemberAddDialog
          organizationId={organizationId}
          roleId={selectedRole.id}
          open={addMemberDialogOpen}
          onOpenChange={setAddMemberDialogOpen}
          onSuccess={() => setAddMemberDialogOpen(false)}
        />
      )}
    </>
  );
}
