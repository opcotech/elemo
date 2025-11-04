import { useQuery } from "@tanstack/react-query";
import { Plus, Trash2 } from "lucide-react";
import { useState } from "react";

import { RoleMemberAddDialog } from "./role-member-add-dialog";
import { RoleMemberRemoveDialog } from "./role-member-remove-dialog";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
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
import type { User } from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import { v1OrganizationRoleMembersGetOptions } from "@/lib/client/@tanstack/react-query.gen";

interface RoleMemberAssignmentProps {
  organizationId: string;
  roleId: string;
  roleName: string;
}

function getInitials(firstName: string, lastName: string): string {
  return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase();
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

  if (isLoading || isPermissionsLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Members</CardTitle>
          <CardDescription>
            Manage members assigned to this role.
          </CardDescription>
        </CardHeader>
        <CardContent>
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
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Members</CardTitle>
          <CardDescription>
            Manage members assigned to this role.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Alert variant="destructive">
            <AlertTitle>Failed to load members</AlertTitle>
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
              <CardTitle>Members</CardTitle>
              <CardDescription>
                Manage members assigned to this role.
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
                Add Member
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          {!members || members.length === 0 ? (
            <div className="text-muted-foreground py-8 text-center text-sm">
              No members assigned to this role yet.
            </div>
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
                {members.map((member) => {
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
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRemoveClick(member)}
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
