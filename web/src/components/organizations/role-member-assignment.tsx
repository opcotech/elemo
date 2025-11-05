import { useQuery } from "@tanstack/react-query";
import { Plus, UserMinus, Users } from "lucide-react";
import { useState } from "react";

import { RoleMemberAddDialog } from "./role-member-add-dialog";
import { RoleMemberRemoveDialog } from "./role-member-remove-dialog";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
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
import type { User } from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import { v1OrganizationRoleMembersGetOptions } from "@/lib/client/@tanstack/react-query.gen";
import { getInitials } from "@/lib/utils";

interface RoleMemberAssignmentProps {
  organizationId: string;
  roleId: string;
  roleName: string;
}

export function RoleMemberAssignment({
  organizationId,
  roleId,
  roleName,
}: RoleMemberAssignmentProps) {
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [removeDialogOpen, setRemoveDialogOpen] = useState(false);
  const [selectedMember, setSelectedMember] = useState<User | null>(null);

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
    data: members,
    isLoading,
    error,
  } = useQuery(
    v1OrganizationRoleMembersGetOptions({
      path: {
        id: organizationId,
        role_id: roleId,
      },
    })
  );

  const handleRemoveClick = (member: User) => {
    setSelectedMember(member);
    setRemoveDialogOpen(true);
  };

  const handleRemoveSuccess = () => {
    setRemoveDialogOpen(false);
    setSelectedMember(null);
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
        Add Member
      </Button>
    ) : undefined;

  const emptyState =
    !members || members.length === 0
      ? {
          icon: <Users />,
          title: "No members assigned",
          description:
            "Add members to grant them the permissions associated with this role.",
          action: hasWritePermission ? (
            <Button
              type="button"
              variant="outline"
              onClick={() => setAddDialogOpen(true)}
              size="sm"
            >
              <Plus className="size-4" />
              Add Member
            </Button>
          ) : undefined,
        }
      : undefined;

  return (
    <>
      <ListContainer
        title="Members"
        description="Manage members assigned to this role."
        isLoading={isLoading || isPermissionsLoading}
        error={error}
        emptyState={emptyState}
        actionButton={createButton}
      >
        {isLoading || isPermissionsLoading ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>
                  <span className="sr-only">Actions</span>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {Array.from({ length: 3 }).map((_, i) => (
                <TableRow key={i}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <Skeleton className="h-10 w-10 rounded-full" />
                      <Skeleton className="h-4 w-32" />
                    </div>
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-40" />
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
                <TableHead>Name</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>
                  <span className="sr-only">Actions</span>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {members?.map((member) => {
                const fullName = `${member.first_name} ${member.last_name}`;
                return (
                  <TableRow key={member.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <Avatar className="h-10 w-10">
                          <AvatarImage
                            src={member.picture || undefined}
                            alt={fullName}
                          />
                          <AvatarFallback>
                            {getInitials(member.first_name, member.last_name)}
                          </AvatarFallback>
                        </Avatar>
                        <span className="font-medium">{fullName}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className="text-muted-foreground text-sm">
                        {member.email}
                      </span>
                    </TableCell>
                    <TableCell className="text-right">
                      {hasWritePermission && (
                        <Button
                          type="button"
                          variant="destructive-ghost"
                          size="sm"
                          onClick={() => handleRemoveClick(member)}
                        >
                          <UserMinus className="h-4 w-4" />
                          <span className="sr-only">Remove member</span>
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

      <RoleMemberAddDialog
        organizationId={organizationId}
        roleId={roleId}
        open={addDialogOpen}
        onOpenChange={setAddDialogOpen}
      />

      {selectedMember && (
        <RoleMemberRemoveDialog
          member={selectedMember}
          roleName={roleName}
          organizationId={organizationId}
          roleId={roleId}
          open={removeDialogOpen}
          onOpenChange={setRemoveDialogOpen}
          onSuccess={handleRemoveSuccess}
        />
      )}
    </>
  );
}
