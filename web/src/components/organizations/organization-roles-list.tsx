import { Link } from "@tanstack/react-router";
import { Edit, Plus, Shield, Trash2, UserPlus } from "lucide-react";
import { useMemo, useState } from "react";

import { OrganizationRoleDeleteDialog } from "./organization-role-delete-dialog";
import { RoleMemberAddDialog } from "./role-member-add-dialog";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ListContainer } from "@/components/ui/list-container";
import { SearchInput } from "@/components/ui/search-input";
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

interface RoleRowProps {
  role: Role;
  organizationId: string;
  onAddMemberClick: (role: Role) => void;
  onDeleteClick: (role: Role) => void;
}

function RoleRow({
  role,
  organizationId,
  onAddMemberClick,
  onDeleteClick,
}: RoleRowProps) {
  const { data: rolePermissions, isLoading: isRolePermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Role, role.id));

  const hasRoleWritePermission = can(rolePermissions, "write");
  const hasRoleDeletePermission = can(rolePermissions, "delete");
  const isPermissionsLoading = isRolePermissionsLoading;

  return (
    <TableRow>
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
  const [searchTerm, setSearchTerm] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [addMemberDialogOpen, setAddMemberDialogOpen] = useState(false);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);

  // Check permissions for organization (read and write)
  const { data: orgPermissions, isLoading: isOrgPermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Organization, organizationId));

  const hasOrgReadPermission = can(orgPermissions, "read");
  const hasOrgWritePermission = can(orgPermissions, "write");
  const hasCreatePermission = hasOrgWritePermission;
  const isPermissionsLoading = isOrgPermissionsLoading;

  // Defense in depth: Don't render if user doesn't have read permission
  if (!isPermissionsLoading && !hasOrgReadPermission) {
    return null;
  }

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

  const filteredRoles = useMemo(() => {
    if (!searchTerm.trim()) return roles;
    const term = searchTerm.toLowerCase();
    return roles.filter(
      (role) =>
        role.name.toLowerCase().includes(term) ||
        (role.description && role.description.toLowerCase().includes(term))
    );
  }, [roles, searchTerm]);

  const createButton =
    !isPermissionsLoading && hasCreatePermission ? (
      <Button variant="outline" size="sm" asChild>
        <Link
          to="/settings/organizations/$organizationId/roles/new"
          params={{ organizationId }}
        >
          <Plus className="size-4" />
          Create Role
        </Link>
      </Button>
    ) : undefined;

  // Only show empty state when there's no data at all (not filtered)
  // When filtered results are empty but original data exists, show search + empty state
  const emptyState =
    roles.length === 0
      ? {
          icon: <Shield />,
          title: "No roles found",
          description:
            "Roles help organize permissions and member access. Create a role to get started.",
          action: hasCreatePermission ? (
            <Button variant="outline" size="sm" asChild>
              <Link
                to="/settings/organizations/$organizationId/roles/new"
                params={{ organizationId }}
              >
                <Plus className="size-4" />
                Create Role
              </Link>
            </Button>
          ) : undefined,
        }
      : filteredRoles.length === 0 && searchTerm.trim()
        ? {
            icon: <Shield />,
            title: "No roles found",
            description:
              "No roles match your search criteria. Try adjusting your search.",
          }
        : undefined;

  // Show search input only when there's data to search through OR when search is active
  const shouldShowSearch = roles.length > 0 || searchTerm.trim() !== "";

  return (
    <>
      <ListContainer
        title="Roles"
        description="Organization roles and their members."
        isLoading={isLoading}
        error={error}
        emptyState={emptyState}
        actionButton={createButton}
        searchInput={
          shouldShowSearch ? (
            <SearchInput
              value={searchTerm}
              onChange={setSearchTerm}
              placeholder="Search roles..."
              disabled={isLoading}
            />
          ) : undefined
        }
      >
        {isLoading ? (
          <OrganizationRolesListSkeleton />
        ) : (
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
              {filteredRoles.map((role) => (
                <RoleRow
                  key={role.id}
                  role={role}
                  organizationId={organizationId}
                  onAddMemberClick={handleAddMemberClick}
                  onDeleteClick={handleDeleteClick}
                />
              ))}
            </TableBody>
          </Table>
        )}
      </ListContainer>

      {selectedRole && (
        <OrganizationRoleDeleteDialog
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
