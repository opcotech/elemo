import { useQuery } from "@tanstack/react-query";
import { Plus, UserMinus, Users } from "lucide-react";
import { useState } from "react";

import { TeamMemberAddDialog } from "./team-member-add-dialog";
import { TeamMemberRemoveDialog } from "./team-member-remove-dialog";

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
import { v1OrganizationTeamMembersGetOptions } from "@/lib/api/query-options";
import type { User } from "@/lib/api/types";
import { Action, can } from "@/lib/auth/permissions";
import { getInitials } from "@/lib/utils";

interface TeamMemberAssignmentProps {
  organizationId: string;
  teamId: string;
  teamName: string;
}

export function TeamMemberAssignment({
  organizationId,
  teamId,
  teamName,
}: TeamMemberAssignmentProps) {
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [removeDialogOpen, setRemoveDialogOpen] = useState(false);
  const [selectedMember, setSelectedMember] = useState<User | null>(null);

  const { data: orgPermissions, isLoading: isOrgPermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Organization, organizationId));

  const hasWritePermission = can(orgPermissions, Action.TeamManage);
  const isPermissionsLoading = isOrgPermissionsLoading;

  const {
    data: membersPage,
    isLoading,
    error,
  } = useQuery(
    v1OrganizationTeamMembersGetOptions({
      path: { organizationRef: organizationId, team_id: teamId },
    })
  );
  const members = membersPage?.items;

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
          description: "Add members so this team can act as a grant principal.",
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
        data-section="team-members"
        title="Members"
        description="Manage members of this team."
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
                          onClick={() => {
                            setSelectedMember(member);
                            setRemoveDialogOpen(true);
                          }}
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

      <TeamMemberAddDialog
        organizationId={organizationId}
        teamId={teamId}
        open={addDialogOpen}
        onOpenChange={setAddDialogOpen}
      />

      {selectedMember && (
        <TeamMemberRemoveDialog
          member={selectedMember}
          teamName={teamName}
          organizationId={organizationId}
          teamId={teamId}
          open={removeDialogOpen}
          onOpenChange={setRemoveDialogOpen}
          onSuccess={() => {
            setRemoveDialogOpen(false);
            setSelectedMember(null);
          }}
        />
      )}
    </>
  );
}
